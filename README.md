# Backend Interview Project

## Testing

I've added integrations tests and two services to the `docker-compose.yml` file: 
- a test database (I use the other database for running the app in dev mode)
- the messaging service, primarily to run migrations on the test database on startup

Once the test database is availabe and the migrations have run, you can run the integration tests using

```bash
make integrations.test
```

I decided to use [gopkg.in/khaiql/dbcleaner.v2](https://github.com/khaiql/dbcleaner) to clean the database after each test run. I probably would not use this in production
because you need to whitelist the database tables that you want cleaned up. Maintain this list of tables may be a problem in the future but I took the opportunity to experiment with it and it does work as specified (for it's defined use case).

## Debugging
I use [delve](https://github.com/go-delve/delve) with my Go projects. I've added a task to run the server with the delve debugger.

```
make debug
```

## Database Indexes
I've added indexes to the database to speed up queries

Based on the `EXPLAIN ANALYZE` output they're working as expected

### Get Conversations

I do have to turn `SET enable_seqscan = OFF;` for _this_ index to be used. There are so few items in the db that the query planner may opt for a sequential scan instead if it's cheaper for this query.
```
Sort  (cost=34.80..34.80 rows=2 width=60) (actual time=1.170..1.351 rows=6 loops=1)
   Sort Key: c.created_at DESC
   Sort Method: quicksort  Memory: 25kB
   ->  Nested Loop Left Join  (cost=17.63..34.79 rows=2 width=60) (actual time=0.416..1.219 rows=6 loops=1)
         ->  Nested Loop Left Join  (cost=17.48..34.39 rows=2 width=24) (actual time=0.363..0.973 rows=6 loops=1)
               ->  Nested Loop  (cost=17.33..33.54 rows=2 width=16) (actual time=0.313..0.717 rows=3 loops=1)
                     ->  Unique  (cost=17.17..17.18 rows=2 width=8) (actual time=0.263..0.559 rows=3 loops=1)
                           ->  Sort  (cost=17.17..17.18 rows=2 width=8) (actual time=0.247..0.350 rows=10 loops=1)
                                 Sort Key: m.conversation_id
                                 Sort Method: quicksort  Memory: 25kB
                                 ->  Bitmap Heap Scan on messages m  (cost=11.82..17.16 rows=2 width=8) (actual time=0.042..0.153 rows=10 loops=1)
                                       Recheck Cond: (message_status = 'success'::message_status)
                                       Heap Blocks: exact=1
                                       ->  Bitmap Index Scan on idx_messages_conversation_id_status  (cost=0.00..11.82 rows=2 width=0) (actual time=0.019..0.026 rows=10 loops=1)
                                             Index Cond: (message_status = 'success'::message_status)
                     ->  Index Scan using conversations_pkey on conversations c  (cost=0.15..8.17 rows=1 width=16) (actual time=0.014..0.016 rows=1 loops=3)
                           Index Cond: (id = m.conversation_id)
               ->  Index Only Scan using conversation_memberships_pkey on conversation_memberships cm  (cost=0.15..0.34 rows=9 width=16) (actual time=0.015..0.031 rows=2 loops=3)
                     Index Cond: (conversation_id = c.id)
                     Heap Fetches: 6
         ->  Index Scan using communications_pkey on communications comm  (cost=0.15..0.20 rows=1 width=44) (actual time=0.012..0.013 rows=1 loops=6)
               Index Cond: (id = cm.communication_id)
 Planning Time: 0.414 ms
 Execution Time: 1.635 ms
```

### Get conversation by ID

```
 Sort  (cost=24.54..24.55 rows=1 width=164) (actual time=0.541..0.646 rows=4 loops=1)
   Sort Key: m.created_at
   Sort Method: quicksort  Memory: 25kB
   ->  Nested Loop Left Join  (cost=0.45..24.53 rows=1 width=164) (actual time=0.195..0.536 rows=4 loops=1)
         ->  Nested Loop  (cost=0.30..16.35 rows=1 width=140) (actual time=0.146..0.285 rows=4 loops=1)
               ->  Index Scan using conversations_pkey on conversations c  (cost=0.15..8.17 rows=1 width=16) (actual time=0.090..0.105 rows=1 loops=1)
                     Index Cond: (id = 1)
               ->  Index Scan using idx_messages_conversation_id_status on messages m  (cost=0.15..8.17 rows=1 width=132) (actual time=0.028..0.066 rows=4 loops=1)
                     Index Cond: ((conversation_id = 1) AND (message_status = 'success'::message_status))
         ->  Index Scan using communications_pkey on communications comm  (cost=0.15..8.17 rows=1 width=40) (actual time=0.013..0.015 rows=1 loops=4)
               Index Cond: (id = m.sender_id)
 Planning Time: 0.320 ms
 Execution Time: 0.790 ms
```


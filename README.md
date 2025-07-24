# Backend Interview Project

## Testing

I've added integrations tests and two services to the `docker-compose.yml` file: 
- a test database (I use the other database for running the app in dev mode)
- the messaging service, primarily to run migrations on the test database on startup

You can run the integration tests using

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
Sort  (cost=34.71..34.72 rows=3 width=48) (actual time=0.821..0.947 rows=3 loops=1)
   Sort Key: c.created_at DESC
   Sort Method: quicksort  Memory: 25kB
   ->  GroupAggregate  (cost=0.41..34.69 rows=3 width=48) (actual time=0.353..0.858 rows=3 loops=1)
         Group Key: c.id
         ->  Nested Loop Left Join  (cost=0.41..34.60 rows=6 width=43) (actual time=0.133..0.629 rows=6 loops=1)
               ->  Merge Left Join  (cost=0.26..24.48 rows=6 width=24) (actual time=0.070..0.287 rows=6 loops=1)
                     Merge Cond: (c.id = cm.conversation_id)
                     ->  Index Scan using conversations_pkey on conversations c  (cost=0.13..12.18 rows=3 width=16) (actual time=0.021..0.051 rows=3 loops=1)
                     ->  Index Only Scan using conversation_memberships_pkey on conversation_memberships cm  (cost=0.13..12.22 rows=6 width=16) (actual time=0.023..0.080 rows=6 loops=1)
                           Heap Fetches: 6
               ->  Memoize  (cost=0.14..2.16 rows=1 width=27) (actual time=0.025..0.029 rows=1 loops=6)
                     Cache Key: cm.communication_id
                     Cache Mode: logical
                     Hits: 1  Misses: 5  Evictions: 0  Overflows: 0  Memory Usage: 1kB
                     ->  Index Scan using communications_pkey on communications comm  (cost=0.13..2.15 rows=1 width=27) (actual time=0.012..0.013 rows=1 loops=5)
                           Index Cond: (id = cm.communication_id)
```

### Get conversation by ID

```
 GroupAggregate  (cost=4.48..26.35 rows=1 width=48) (actual time=0.706..0.811 rows=1 loops=1)
   Group Key: c.id
   ->  Nested Loop Left Join  (cost=4.48..26.33 rows=1 width=164) (actual time=0.229..0.521 rows=4 loops=1)
         ->  Nested Loop Left Join  (cost=4.32..19.49 rows=1 width=140) (actual time=0.182..0.327 rows=4 loops=1)
               Join Filter: (m.conversation_id = c.id)
               ->  Index Scan using conversations_pkey on conversations c  (cost=0.15..8.17 rows=1 width=16) (actual time=0.126..0.140 rows=1 loops=1)
                     Index Cond: (id = 2)
               ->  Bitmap Heap Scan on messages m  (cost=4.17..11.28 rows=3 width=132) (actual time=0.028..0.076 rows=4 loops=1)
                     Recheck Cond: (conversation_id = 2)
                     Heap Blocks: exact=1
                     ->  Bitmap Index Scan on idx_messages_conversation_id_created_at  (cost=0.00..4.17 rows=3 width=0) (actual time=0.011..0.018 rows=4 loops=1)
                           Index Cond: (conversation_id = 2)
         ->  Index Scan using communications_pkey on communications comm  (cost=0.15..6.84 rows=1 width=40) (actual time=0.013..0.015 rows=1 loops=4)
               Index Cond: (id = m.sender_id)
```


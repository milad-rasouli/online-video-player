## Here you will see how I used Redis to store chat message:

```bash
# User1 sends a message
HSET message:1 user_id "user1" message_text "Hello, world!" timestamp "1620138245"
EXPIRE message:1 60
LPUSH messages message:1

# User2 sends a message
HSET message:2 user_id "user2" message_text "Hi, everyone!" timestamp "1620138345"
EXPIRE message:2 60
LPUSH messages message:2

# A new user joins the chat and retrieves the messages
SORT messages BY *->timestamp GET *->user_id GET *->message_text
```

> [! TIP]
> set expire time `EXPIRE message:1 1`

## LUA scripting

### Save a message in Redis:

```bash
SCRIPT LOAD "redis.call('HSET', KEYS[1], 'user_id', ARGV[1], 'message_text', ARGV[2], 'timestamp', ARGV[3]) redis.call('EXPIRE', KEYS[1], 60) redis.call('LPUSH', 'messages', KEYS[1])"

"811c8cd082f7823c009788a36b9b407bb0cdf725"
```

### Inserting some dummy message

```bash
EVALSHA "811c8cd082f7823c009788a36b9b407bb0cdf725" 1 message:2 user2 "Hi, everyone!" 1620138345
EVALSHA "811c8cd082f7823c009788a36b9b407bb0cdf725" 1 message:3 user3 "Good afternoon!" 1620138445
EVALSHA "811c8cd082f7823c009788a36b9b407bb0cdf725" 1 message:1 user1 "Hello, world!" 1620138245

```

## Retuning message

```bash

SCRIPT LOAD " local messages = redis.call('LRANGE', KEYS[1], 0, -1) local valid_messages = {} for i, message_key in ipairs(messages) do if redis.call('EXISTS', message_key) == 1 then local message = redis.call('HGETALL', message_key) table.insert(valid_messages, message) else redis.call('LREM', KEYS[1], 0, message_key) end end return valid_messages "

"ad22087b1a20fa647e991e5a566c31fdf919fbfb"
```

```bash
EVALSHA "ad22087b1a20fa647e991e5a566c31fdf919fbfb" 1 messages
```

## How to online users are stored

### Inserting users

```bash
SET user:name1 value EX 500
SET user:name2 value EX 500
SET user:name3 value EX 500
```

### Getting users

```bash
SCAN 0 MATCH user:*
```

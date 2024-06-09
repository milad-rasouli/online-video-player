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
>[! TIP]
> set expire time ``` EXPIRE message:1 1 ```

## LUA scripting

To add a chat message:
```bash
-- The script takes 3 arguments: message_id, user_id, and message_text
local message_id = KEYS[1]
local user_id = ARGV[1]
local message_text = ARGV[2]

-- Get the current time in seconds (Unix timestamp)
local time = redis.call('TIME')
local timestamp = time[1]

-- Store the message details as a hash
redis.call('HSET', message_id, 'user_id', user_id, 'message_text', message_text, 'timestamp', timestamp)

-- Set an expiration time of 60 seconds for the message
redis.call('EXPIRE', message_id, 60)

-- Add the message_id to a list
redis.call('LPUSH', 'messages', message_id)



-- The script takes 3 arguments: message_id, user_id, and message_text
local message_id = KEYS[1]
local user_id = ARGV[1]
local message_text = ARGV[2]

-- Get the current time in seconds (Unix timestamp)
local time = redis.call('TIME')
local timestamp = time[1]

-- Store the message details as a hash
redis.call('HSET', message_id, 'user_id', user_id, 'message_text', message_text, 'timestamp', timestamp)

-- Set an expiration time of 60 seconds for the message
redis.call('EXPIRE', message_id, 60)

redis.call('LPUSH', 'messages', message_id)
```


```bash
SCRIPT LOAD "local message_id = KEYS[1] local user_id = ARGV[1] local message_text = ARGV[2] local time = redis.call('TIME') local timestamp = time[1] redis.call('HSET', message_id, 'user_id', user_id, 'message_text', message_text, 'timestamp', timestamp) redis.call('EXPIRE', message_id, 60) redis.call('LPUSH', 'messages', message_id)"

"3e8d55799190514aae47977749f3c7c53b40e4df"
```

```bash 
EVALSHA "3e8d55799190514aae47977749f3c7c53b40e4df" 1 message:10 user1 "hello world!"
```


## Clean up nil parameters

```bash
local messages = redis.call('LRANGE', 'messages', 0, -1)
for i, message_id in ipairs(messages) do
    if not redis.call('EXISTS', message_id) then
        redis.call('LREM', 'messages', 0, message_id)
    end
end
```

```bash
SCRIPT LOAD "local messages = redis.call('LRANGE', 'messages', 0, -1) for i, message_id in ipairs(messages) do if not redis.call('EXISTS', message_id) then redis.call('LREM', 'messages', 0, message_id) end end"

"3c514d4d34fe788fff4fef9c04ec7aa27fc8b19c"
```


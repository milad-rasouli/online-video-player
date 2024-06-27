# online-video-player
A website that helps you to watch videos with your friends and chat with them.
here
[https://github.com/users/Milad75Rasouli/projects/10Here is where you can see my plans for the project.](https://github.com/users/Milad75Rasouli/projects/10) If you have any suggestions to enhance this project, please feel free to open an issue.
## How to run the program

1. Run redis:
```bash
sudo docker compose up -d
```
2- Run the program:
```bash
just run
```

> [!NOTE]
> if you have trouble running with *just* you should install it on your machine first. [Check this out](https://github.com/casey/just)

## How to run the project with Docker

```bash
# docker run --init --rm --name player -e USER_PASSWORD="1234qwer" -e PROGRAM_PORT=":5000" -e DEBUG="false" -e WEBSITE_TITLE="Online Player" -e REDIS_ADDRESS="127.0.0.1:6379" -e REDIS_CHAT_EXP="60" -e USER_PASSWORD="123" -e JWT_SECRET="changeIt" -e JWT_EXPIRE_TIME="3" -p 5000:5000 ghcr.io/milad75rasouli/player:latest

```

## How to use the program

1. Go to http://localhost:5000/auth then enter your name and password(default password is *123*)
2. Put the video download link in the *Video URL* and hit the *submit*. If everything goes okay you should see the download precess. Also you can cancel the process whenever you feel like.
3. After the download process gets completed you're able to watch it with your friends and chat with them at the same time.

C## Contribution
Your contributions to this project are welcome. Please feel free to open issues and send pull requests.


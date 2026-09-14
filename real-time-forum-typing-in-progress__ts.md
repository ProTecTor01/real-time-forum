## typing in progress

### Objectives

You must follow the same [principles](https://01.tomorrow-school.ai/api/content/root/public/subjects/real-time-forum/README.md) as the first subject.

For this project you must create:

- A [Typing in progress](https://i.insider.com/56996788dd0895a06c8b460c?width=1100&format=jpeg&auto=webp) engine.

### Instructions

A typing in progress engine is a way that people can see that a user is typing in real time. Allowing you to see the other user is replying or sending a message.

The typing in progress engine must work in real time! This meaning that if you start typing to a certain user this user will be able to see that you are typing.

This engine must have/display:

- A websocket to establish the connection with both users
- An animation so that the user can see that you are typing, this animation should be smooth (no interruptions/janks) and just enough to draw attention for the user to see (user friendly)
- The name of the user that is typing
- Whenever the user stops typing or finishes the conversation, it should not display the animation, a short delay may optionally be added for better UX

To help with the display of the typing in progress you can take a look on the js [event](https://developer.mozilla.org/en-US/docs/Web/Events) list, mainly the **Keyboard events** and the **Focus events**

### Allowed Packages

- All [standard go](https://golang.org/pkg/) packages are allowed.
- [Gorilla websocket](https://pkg.go.dev/github.com/gorilla/websocket)
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

This project will help you learn about:

- [Go routines](https://golangbot.com/goroutines/)
- [Go channels](https://medium.com/rungo/anatomy-of-channels-in-go-concurrency-in-go-1ec336086adb)
- [WebSockets](https://en.wikipedia.org/wiki/WebSocket):
  - Go Websockets
  - JS Websockets
- [Events](https://developer.mozilla.org/en-US/docs/Web/Events)






========== AUDIT ==========





#### Functional

###### Has the requirement for the allowed packages been respected? (Check the [allowed packages](https://01.tomorrow-school.ai/api/content/root/public/subjects/real-time-forum/typing-in-progress/audit/README.md))

##### Open two browsers (ex: Chrome and Firefox or private windows) and log in with different users in each. Start typing with one of the users.

###### Is there any animation confirming that the typing in progress is working?

##### Using the same two browsers, start typing with one of the users.

###### Does the animation work smoothly, without movement interruptions?

##### Using the same two browsers, start typing with one of the users.

###### Is the animation from the typing in progress engine user friendly (easy to understand/see)?

##### Using the same two browsers, start typing with one of the users.

###### Can you confirm that the typing in progress has the name of the user that is typing?

##### Using the same two browsers, start typing with one of the users, then stop typing.

###### Can you confirm that the typing in progress engine stopped when the user stop typing?

##### Using the same two browsers, start a conversation between two users.

###### Is the typing in progress engine working properly for both users? (each one should be able to see whether the other is typing or not)

#### Bonus

###### +Does the project run quickly and effectively? (Favoring recursivity, no unnecessary data requests, etc...)

###### +Does the code obey the [good practices](https://01.tomorrow-school.ai/api/content/root/public/subjects/good-practices/README.md)?

###### +Is the code using synchronicity (Promises and Go routines/channels) to increase performance?

###### +Do you think this project is well done in general ?
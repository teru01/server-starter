# brief description

`server-starter` restarts server process without disconnecting. It creates a socket listening on specified ports, then executes the given command as sub process, passing the created socket. When this process receives SIGHUP, it sends SIGTERM to the sub process and executes command again.

package pkg

// Announcer ...
type Announcer interface {
	Announce(string, string, *DockerTag)
}

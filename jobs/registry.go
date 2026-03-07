package jobs

import "github.com/mahavirnahata/gophant/queue"

// Registry is a global job registry apps can use for queue workers.
var Registry = queue.NewRegistry()

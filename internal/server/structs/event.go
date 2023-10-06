package structs

// EventTopic is a string type that represents a topic for an event.
type EventTopic string

const (
	// EventTopicNode is the topic for Node events.
	EventTopicNode EventTopic = "Node"

	// EventTypeNodeCreated is the type for Node created events.
	EventTypeNodeCreated = "NodeCreated"
	// EventTypeNodeUpdated is the type for Node updated events.
	EventTypeNodeUpdated = "NodeUpdated"
	// EventTypeNodeDeleted is the type for Node deleted events.
	EventTypeNodeDeleted = "NodeDeleted"
)

// Event is a struct that represents an event that happened in the system.
type Event struct {
	Topic EventTopic
	Type  string

	// Object is the object that was created, updated, deleted or whatever.
	Object interface{}
}

// NewNodeEvent creates a new Event of Node topic with the given type and Node object.
func NewNodeEvent(t string, n *Node) *Event {
	e := &Event{
		Topic:  EventTopicNode,
		Type:   t,
		Object: n,
	}

	return e
}

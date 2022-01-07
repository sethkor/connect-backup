package connect_backup

var defaultFlows = map[string]bool{
	"Sample inbound flow (first contact experience)": true,
	"Default agent hold":                             true,
	"Default customer queue":                         true,
	"Default agent whisper":                          true,
	"Import Testing":                                 true,
	"Sample note for screenpop":                      true,
	"Sample AB test":                                 true,
	"Default queue transfer":                         true,
	"Sample disconnect flow":                         true,
	"Sample queue configurations flow":               true,
	"Default customer hold":                          true,
	"Default agent transfer":                         true,
	"Default outbound":                               true,
	"Sample recording behavior":                      true,
	"Sample Lambda integration":                      true,
	"Sample queue customer":                          true,
	"Sample secure input with no agent":              true,
	"Default customer whisper":                       true,
	"Sample interruptible queue flow with callback":  true,
	"Sample secure input with agent":                 true,
}

type ConnectElement string

const (
	Flows                  ConnectElement = "flows"
	FlowsRaw               ConnectElement = "flows-raw"
	RoutingProfiles        ConnectElement = "routing-profiles"
	RoutingProfileQueues   ConnectElement = "routing-profile-queues"
	Users                  ConnectElement = "users"
	UserHierarchyGroups    ConnectElement = "user-hierarchy-groups"
	UserHierarchyStructure ConnectElement = "user-hierarchy-structures"
	Prompts                ConnectElement = "prompts"
	HoursOfOperation       ConnectElement = "hours-of-operation"
	QuickConnects          ConnectElement = "quick-connects"
	Queues                 ConnectElement = "queues"
	Instance               ConnectElement = "instance"
)

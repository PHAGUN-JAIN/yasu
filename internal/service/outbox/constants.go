package outbox

const (	
	defaultBatchSize = 1000
	defaultFlushEvery = 30 
	defaultPollEvery = 30 
)

const (
	SourceREST string = "rest"
	SourceMCP  string = "mcp"
	SourceA2A  string = "a2a"
)

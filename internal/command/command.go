package command

type Command struct {
	Action Action `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
}

type Commands []Command

func NewSetCommand(key, val string) Command {
	return Command{
		Action: SetAction,
		Key:    key,
		Value:  val,
	}
}

func NewDeleteCommand(key string) Command {
	return Command{
		Action: DeleteAction,
		Key:    key,
	}
}

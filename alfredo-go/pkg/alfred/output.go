package alfred

import "encoding/json"

// Output represents the Alfred workflow output format
type Output struct {
	Items []OutputItem `json:"items"`
}

// OutputItem represents a single item in Alfred workflow output
type OutputItem struct {
	Title     string              `json:"title"`
	Subtitle  string              `json:"subtitle"`
	Arg       string              `json:"arg"`
	Variables map[string]any      `json:"variables,omitempty"`
	Mods      map[string]ModsItem `json:"mods,omitempty"`
	Icon      *Icon               `json:"icon,omitempty"`
}

// ModsItem represents modifier keys in Alfred workflow
type ModsItem struct {
	Arg      string `json:"arg"`
	Subtitle string `json:"subtitle"`
}

// Icon represents an Alfred item icon
type Icon struct {
	Path string `json:"path"`
}

// Marshal returns the JSON representation of the output
func (o *Output) Marshal() ([]byte, error) {
	return json.Marshal(o)
}

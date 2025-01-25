package helperapi

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// JsonPatch represents an RFC6902 compilant JSON patch
type JsonPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// Applies a sequence of [JsonPatch] patches to a provided map, in place.
// Returns an error if the patch operation fails.
func (api *Api) ApplyJsonPatches(to any, patches ...JsonPatch) error {
	if (reflect.ValueOf(to).Kind() != reflect.Ptr) {
		return fmt.Errorf("to must be pointer")
	}
	api.Logger.Info("apply json patches", "count", len(patches))
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		return err
	}

	toBytes, err := json.Marshal(to)
	if err != nil {
		return err
	}

	jsonPatch, err := jsonpatch.DecodePatch(patchesBytes)
	if err != nil {
		return err
	}

	toBytes, err = jsonPatch.ApplyIndent(toBytes, "  ")
	if err != nil {
		return err
	}

	err = json.Unmarshal(toBytes, &to)
	if err != nil {
		return err
	}

	return nil
}

package parser

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// QueryParseStep parses parameters from the query string of a request.
type QueryParseStep struct {
	onlyProvided   bool
	ignoreRequired bool
}

// TODO(manuel, 2023-10-17) We need to figure out how to figure out that missing parameters are required
// This is where we need to collect all validations (not just missing, but invalid) so that we can always
// display the form filled with values

func (q *QueryParseStep) ParseLayerState(c *gin.Context, state *LayerParseState) error {
	for _, p := range state.ParameterDefinitions {
		if parameters.IsListParameter(p.Type) {
			// check p.Name[] parameter
			values, ok := c.GetQueryArray(fmt.Sprintf("%s[]", p.Name))
			if ok {
				pValue, err := p.ParseParameter(values)
				if err != nil {
					return fmt.Errorf("invalid value for parameter '%s': (%v) %s", p.Name, values, err.Error())
				}
				state.Parameters[p.Name] = pValue
				continue
			}
			if p.Required {
				if q.ignoreRequired {
					continue
				}
				return errors.Errorf("required parameter '%s' is missing", p.Name)
			}
		}
		value := c.DefaultQuery(p.Name, state.Defaults[p.Name])
		if parameters.IsFileLoadingParameter(p.Type, value) {
			// if the parameter is supposed to be read from a file, we will just pass in the query parameters
			// as a placeholder here
			if value == "" {
				if p.Required {
					if q.ignoreRequired {
						continue
					}
					return errors.Errorf("required parameter '%s' is missing", p.Name)
				}
				if !q.onlyProvided {
					if _, ok := state.Parameters[p.Name]; !ok {
						state.Parameters[p.Name] = p.Default
					}
				}
			} else {
				f := strings.NewReader(value)
				pValue, err := p.ParseFromReader(f, "")
				if err != nil {
					return fmt.Errorf("invalid value for parameter '%s': (%v) %s", p.Name, value, err.Error())
				}
				state.Parameters[p.Name] = pValue
			}
		} else {
			if value == "" {
				if p.Required {
					if q.ignoreRequired {
						continue
					}
					return fmt.Errorf("required parameter '%s' is missing", p.Name)
				}
				if !q.onlyProvided {
					if p.Type == parameters.ParameterTypeDate {
						switch v := p.Default.(type) {
						case string:
							parsedDate, err := parameters.ParseDate(v)
							if err != nil {
								return fmt.Errorf("invalid value for parameter '%s': (%v) %s", p.Name, value, err.Error())
							}

							state.Parameters[p.Name] = parsedDate
						case time.Time:
							state.Parameters[p.Name] = v

						}
					} else {
						// only set default value if it is not already set
						if _, ok := state.Parameters[p.Name]; !ok {
							state.Parameters[p.Name] = p.Default
						}
					}
				}
			} else {
				var values []string
				if parameters.IsListParameter(p.Type) {
					values = strings.Split(value, ",")
				} else {
					values = []string{value}
				}
				pValue, err := p.ParseParameter(values)
				if err != nil {
					return fmt.Errorf("invalid value for parameter '%s': (%v) %s", p.Name, value, err.Error())
				}
				state.Parameters[p.Name] = pValue
			}
		}
	}

	return nil
}

func (q *QueryParseStep) Parse(c *gin.Context, state *LayerParseState) error {
	err := q.ParseLayerState(c, state)
	if err != nil {
		return err
	}

	return nil
}

func NewQueryParseStep(onlyProvided bool, ignoreRequired bool) ParseStep {
	return &QueryParseStep{
		onlyProvided:   onlyProvided,
		ignoreRequired: ignoreRequired,
	}
}

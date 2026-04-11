// Package sdk provides types for defining Yeetcd pipelines in Go.
package sdk

import (
	"fmt"
	"os"
	"reflect"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

// Pipeline represents a complete pipeline definition.
type Pipeline struct {
	Name        string
	Parameters  Parameters
	WorkContext WorkContext
	FinalWork   []Work
}

// PipelineBuilder builds a Pipeline.
type PipelineBuilder struct {
	pipeline Pipeline
}

// NewPipeline creates a new PipelineBuilder.
func NewPipeline(name string) *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: Pipeline{
			Name:        name,
			Parameters:  Parameters{},
			WorkContext: WorkContext{},
			FinalWork:   []Work{},
		},
	}
}

// WithParameters sets the parameters.
func (b *PipelineBuilder) WithParameters(params Parameters) *PipelineBuilder {
	b.pipeline.Parameters = params
	return b
}

// WithWorkContext sets the work context.
func (b *PipelineBuilder) WithWorkContext(ctx WorkContext) *PipelineBuilder {
	b.pipeline.WorkContext = ctx
	return b
}

// WithFinalWork sets the final work.
func (b *PipelineBuilder) WithFinalWork(work ...Work) *PipelineBuilder {
	b.pipeline.FinalWork = work
	return b
}

// Build returns the Pipeline.
func (b *PipelineBuilder) Build() Pipeline {
	return b.pipeline
}

// Work represents a unit of work in a pipeline.
type Work struct {
	Description    string
	WorkDefinition WorkDefinition
	WorkContext    WorkContext
	OutputPaths    []WorkOutputPath
	PreviousWork   []PreviousWork
	Condition      Condition
}

// WorkBuilder builds a Work.
type WorkBuilder struct {
	work Work
}

// NewWork creates a new WorkBuilder.
func NewWork(description string, workDef WorkDefinition) *WorkBuilder {
	return &WorkBuilder{
		work: Work{
			Description:    description,
			WorkDefinition: workDef,
			WorkContext:    WorkContext{},
			OutputPaths:    []WorkOutputPath{},
			PreviousWork:   []PreviousWork{},
			Condition:      nil,
		},
	}
}

// WithWorkContext sets the work context.
func (b *WorkBuilder) WithWorkContext(ctx WorkContext) *WorkBuilder {
	b.work.WorkContext = ctx
	return b
}

// WithOutputPaths sets output paths.
func (b *WorkBuilder) WithOutputPaths(paths ...WorkOutputPath) *WorkBuilder {
	b.work.OutputPaths = paths
	return b
}

// WithPreviousWork sets previous work.
func (b *WorkBuilder) WithPreviousWork(work ...PreviousWork) *WorkBuilder {
	b.work.PreviousWork = work
	return b
}

// WithCondition sets a condition.
func (b *WorkBuilder) WithCondition(cond Condition) *WorkBuilder {
	b.work.Condition = cond
	return b
}

// Build returns the Work.
func (b *WorkBuilder) Build() Work {
	return b.work
}

// WorkContext holds key-value pairs passed to work as environment variables.
type WorkContext map[string]string

// EmptyWorkContext returns an empty work context.
func EmptyWorkContext() WorkContext {
	return WorkContext{}
}

// WorkContextOf creates a work context from key-value pairs.
func WorkContextOf(pairs ...string) WorkContext {
	if len(pairs)%2 != 0 {
		panic("WorkContextOf requires even number of arguments")
	}
	ctx := make(WorkContext)
	for i := 0; i < len(pairs); i += 2 {
		ctx[pairs[i]] = pairs[i+1]
	}
	return ctx
}

// Merge merges this context with another, with this context taking precedence.
func (c WorkContext) Merge(other WorkContext) WorkContext {
	result := make(WorkContext)
	for k, v := range other {
		result[k] = v
	}
	for k, v := range c {
		result[k] = v
	}
	return result
}

// WorkContextValue retrieves a value from work context (environment variable).
func WorkContextValue(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return ""
}

// WorkOutputPath represents a path for work output.
type WorkOutputPath struct {
	Name string
	Path string
}

// NewWorkOutputPath creates a new WorkOutputPath.
func NewWorkOutputPath(name, path string) WorkOutputPath {
	return WorkOutputPath{Name: name, Path: path}
}

// PreviousWork represents a dependency on a previous work item.
type PreviousWork struct {
	Work             Work
	OutputPathsMount string
	StdOutEnvVar     string
}

// PreviousWorkBuilder builds a PreviousWork.
type PreviousWorkBuilder struct {
	pw PreviousWork
}

// NewPreviousWork creates a new PreviousWorkBuilder.
func NewPreviousWork(work Work) *PreviousWorkBuilder {
	return &PreviousWorkBuilder{
		pw: PreviousWork{Work: work},
	}
}

// WithOutputsMountPath sets the outputs mount path.
func (b *PreviousWorkBuilder) WithOutputsMountPath(path string) *PreviousWorkBuilder {
	b.pw.OutputPathsMount = path
	return b
}

// WithStdOutEnvVar sets the stdout environment variable.
func (b *PreviousWorkBuilder) WithStdOutEnvVar(name string) *PreviousWorkBuilder {
	b.pw.StdOutEnvVar = name
	return b
}

// Build returns the PreviousWork.
func (b *PreviousWorkBuilder) Build() PreviousWork {
	return b.pw
}

// WorkDefinition is the interface that all work definitions must implement.
type WorkDefinition interface {
	// toProto converts the work definition to protobuf format
	toProto() proto.Message
}

// ContainerisedWorkDefinition runs a command in an existing container image.
type ContainerisedWorkDefinition struct {
	Image string
	Cmd   []string
}

// NewContainerisedWork creates a ContainerisedWorkDefinition.
func NewContainerisedWork(image string) *ContainerisedWorkDefinitionBuilder {
	return &ContainerisedWorkDefinitionBuilder{
		def: &ContainerisedWorkDefinition{Image: image},
	}
}

// ContainerisedWorkDefinitionBuilder builds a ContainerisedWorkDefinition.
type ContainerisedWorkDefinitionBuilder struct {
	def *ContainerisedWorkDefinition
}

// WithCommand sets the command.
func (b *ContainerisedWorkDefinitionBuilder) WithCommand(cmd ...string) *ContainerisedWorkDefinitionBuilder {
	b.def.Cmd = cmd
	return b
}

// Build returns the ContainerisedWorkDefinition.
func (b *ContainerisedWorkDefinitionBuilder) Build() *ContainerisedWorkDefinition {
	return b.def
}

func (c *ContainerisedWorkDefinition) toProto() proto.Message {
	return nil // Placeholder - will use generated protobuf
}

// CustomWorkDefinition is a user-defined work that runs custom Go code.
type CustomWorkDefinition struct {
	execute func()
}

// NewCustomWork creates a CustomWorkDefinition.
func NewCustomWork(execute func()) *CustomWorkDefinitionBuilder {
	return &CustomWorkDefinitionBuilder{
		def: &CustomWorkDefinition{execute: execute},
	}
}

// CustomWorkDefinitionBuilder builds a CustomWorkDefinition.
type CustomWorkDefinitionBuilder struct {
	def *CustomWorkDefinition
}

// Build returns the CustomWorkDefinition.
func (b *CustomWorkDefinitionBuilder) Build() *CustomWorkDefinition {
	return b.def
}

// Run executes the custom work.
func (c *CustomWorkDefinition) Run() {
	if c.execute != nil {
		c.execute()
	}
}

func (c *CustomWorkDefinition) toProto() proto.Message {
	return nil // Placeholder
}

func (c *CustomWorkDefinition) executionID() string {
	return fmt.Sprintf("%x", reflect.TypeOf(c).String())
}

// CompoundWorkDefinition groups multiple work items as a single unit.
type CompoundWorkDefinition struct {
	FinalWork []Work
}

// NewCompoundWork creates a CompoundWorkDefinition.
func NewCompoundWork(work ...Work) *CompoundWorkDefinitionBuilder {
	return &CompoundWorkDefinitionBuilder{
		def: &CompoundWorkDefinition{FinalWork: work},
	}
}

// CompoundWorkDefinitionBuilder builds a CompoundWorkDefinition.
type CompoundWorkDefinitionBuilder struct {
	def *CompoundWorkDefinition
}

// Build returns the CompoundWorkDefinition.
func (b *CompoundWorkDefinitionBuilder) Build() *CompoundWorkDefinition {
	return b.def
}

func (c *CompoundWorkDefinition) toProto() proto.Message {
	return nil // Placeholder
}

// DynamicWorkGeneratingWorkDefinition generates work at runtime.
type DynamicWorkGeneratingWorkDefinition struct {
	Generate func() []Work
}

// NewDynamicWork creates a DynamicWorkGeneratingWorkDefinition.
func NewDynamicWork(generate func() []Work) *DynamicWorkGeneratingWorkDefinition {
	return &DynamicWorkGeneratingWorkDefinition{Generate: generate}
}

func (d *DynamicWorkGeneratingWorkDefinition) toProto() proto.Message {
	return nil // Placeholder
}

// Condition is the interface for work execution conditions.
type Condition interface {
	toProto() proto.Message
}

// WorkContextCondition checks work context values.
type WorkContextCondition struct {
	Key     string
	Operand Operand
	Value   string
}

// Operand for work context conditions.
type Operand string

const (
	OperandEquals Operand = "EQUALS"
)

// NewWorkContextCondition creates a WorkContextCondition.
func NewWorkContextCondition(key string, operand Operand, value string) *WorkContextConditionBuilder {
	return &WorkContextConditionBuilder{
		cond: &WorkContextCondition{Key: key, Operand: operand, Value: value},
	}
}

// WorkContextConditionBuilder builds a WorkContextCondition.
type WorkContextConditionBuilder struct {
	cond *WorkContextCondition
}

// Build returns the WorkContextCondition.
func (b *WorkContextConditionBuilder) Build() *WorkContextCondition {
	return b.cond
}

func (c *WorkContextCondition) toProto() proto.Message {
	return nil // Placeholder
}

// NotCondition inverts another condition.
type NotCondition struct {
	Condition Condition
}

// AndCondition combines two conditions with AND.
type AndCondition struct {
	Left  Condition
	Right Condition
}

// OrCondition combines two conditions with OR.
type OrCondition struct {
	Left  Condition
	Right Condition
}

// PreviousWorkStatusCondition checks previous work status.
type PreviousWorkStatusCondition struct {
	Status Status
}

// Status of previous work.
type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusFailure Status = "FAILURE"
	StatusAny     Status = "ANY"
)

// Conditions provides helper functions for creating conditions.
var Conditions = struct {
	WorkContextCondition func(key string, operand Operand, value string) *WorkContextCondition
	Not                  func(Condition) *NotCondition
	And                  func(Condition, Condition) *AndCondition
	Or                   func(Condition, Condition) *OrCondition
	PreviousWorkStatus   func(Status) *PreviousWorkStatusCondition
}{
	WorkContextCondition: func(key string, operand Operand, value string) *WorkContextCondition {
		return &WorkContextCondition{Key: key, Operand: operand, Value: value}
	},
	Not: func(c Condition) *NotCondition {
		return &NotCondition{Condition: c}
	},
	And: func(left, right Condition) *AndCondition {
		return &AndCondition{Left: left, Right: right}
	},
	Or: func(left, right Condition) *OrCondition {
		return &OrCondition{Left: left, Right: right}
	},
	PreviousWorkStatus: func(status Status) *PreviousWorkStatusCondition {
		return &PreviousWorkStatusCondition{Status: status}
	},
}

func (n *NotCondition) toProto() proto.Message                { return nil }
func (a *AndCondition) toProto() proto.Message                { return nil }
func (o *OrCondition) toProto() proto.Message                 { return nil }
func (p *PreviousWorkStatusCondition) toProto() proto.Message { return nil }

// Pipelines is a collection of pipeline definitions.
type Pipelines []Pipeline

// Parameters holds parameter definitions.
type Parameters map[string]Parameter

// ParametersOf creates parameters from key-value pairs.
func ParametersOf(pairs ...interface{}) Parameters {
	if len(pairs)%2 != 0 {
		panic("ParametersOf requires even number of arguments")
	}
	params := make(Parameters)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic("parameter key must be string")
		}
		params[key] = pairs[i+1].(Parameter)
	}
	return params
}

// EmptyParameters returns an empty parameters map.
func EmptyParameters() Parameters {
	return Parameters{}
}

// Parameter defines a pipeline parameter.
type Parameter struct {
	TypeCheck    TypeCheck
	Required     bool
	DefaultValue string
	Choices      []string
}

// TypeCheck defines the type validation for a parameter.
type TypeCheck string

const (
	TypeCheckString  TypeCheck = "STRING"
	TypeCheckNumber  TypeCheck = "NUMBER"
	TypeCheckBoolean TypeCheck = "BOOLEAN"
)

// NewParameter creates a new ParameterBuilder.
func NewParameter(typeCheck TypeCheck) *ParameterBuilder {
	return &ParameterBuilder{
		param: Parameter{TypeCheck: typeCheck},
	}
}

// ParameterBuilder builds a Parameter.
type ParameterBuilder struct {
	param Parameter
}

// WithRequired sets required.
func (b *ParameterBuilder) WithRequired(required bool) *ParameterBuilder {
	b.param.Required = required
	return b
}

// WithDefaultValue sets default value.
func (b *ParameterBuilder) WithDefaultValue(value string) *ParameterBuilder {
	b.param.DefaultValue = value
	return b
}

// WithChoices sets choices.
func (b *ParameterBuilder) WithChoices(choices ...string) *ParameterBuilder {
	b.param.Choices = choices
	return b
}

// Build returns the Parameter.
func (b *ParameterBuilder) Build() Parameter {
	return b.param
}

// WorkID generates a unique ID for a work item.
var workIDCounter atomic.Int64

func generateWorkID() string {
	return fmt.Sprintf("work-%d", workIDCounter.Add(1))
}

// PipelineRunner is the interface for running pipelines.
// This will be implemented by the test SDK.
type PipelineRunner interface {
	Run(pipelineName string, args map[string]string) (*PipelineRunResult, error)
}

// PipelineRunResult represents the result of a pipeline run.
type PipelineRunResult struct {
	Status         string
	ExitCode       int
	WorkExecutions []WorkExecution
}

// WorkExecution represents a single work execution.
type WorkExecution struct {
	WorkID      string
	Description string
	Status      string
}

// MockWorkBehavior defines behavior for mocking work execution.
type MockWorkBehavior struct {
	Execute func() error
}

// NewMockWorkBehavior creates a MockWorkBehavior.
func NewMockWorkBehavior(execute func() error) *MockWorkBehavior {
	return &MockWorkBehavior{Execute: execute}
}

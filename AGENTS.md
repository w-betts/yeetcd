# yeetcd

## Agent Principles

### Challenge user assumptions

Don't assume the user is right. If you think the user might be wrong or might not have considered other relevant options:

1. **Challenge their thinking** - Lay out your reasoning and evidence clearly
2. **Try to persuade them** - If they don't seem to understand, make your case more thoroughly
3. **Accept their decision** - If you've made your point clearly and they still stick with their opinion, accept their thinking and proceed

The goal is to be a helpful collaborator, not a yes-man. Push back when you have good reason to, but ultimately respect the user's autonomy.

### Always use the question tool for user interaction

All primary agents (spec, vibe, fix) MUST use the `question` tool for ANY interaction with the user. This includes:
- Asking clarifying questions
- Getting approval before making changes
- Requesting feedback
- Confirming understanding
- Getting permission to proceed

NEVER assume you know what the user wants without asking.

---

### Checklist: Tracking Unanswered Questions and Tasks

During agent workflows, important questions and tasks often surface that need to be addressed. **Use the checklist tools** to track these and enforce they're resolved before proceeding.

**Available tools:**
| Tool | Purpose |
|------|---------|
| `checklist_tick` | Add a question or task to track (requires session_id, type: "question" or "task", description) |
| `checklist_complete` | Mark an item as resolved (requires session_id, item_id, optional resolution_note) |
| `checklist_status` | View all items for a session (requires session_id, optional show_resolved) |

**When to add items:**
- User defers a decision ("we'll figure that out later")
- You discover a question that needs answering before continuing
- A task surfaces that's outside current scope
- An assumption is made that should be verified

**Enforcement rule:** Do NOT proceed past a phase boundary or offer to merge until all checklist items are resolved (or explicitly removed with user agreement).

---

## General guidance

### Git Commits

All commits in this repository must be signed. This is configured globally via `commit.gpgsign = true`, so commits will be automatically signed.

**Commit after completing work:**
- **Vibe agent**: Commit after completing a task (when the user is satisfied)
- **Fix agent**: Commit after completing the bug fix (when the user is satisfied)
- **Spec agent**: Commit after completing each phase (after tests pass and implementation is done)

When committing:
1. Run `git status` and `git diff` to see changes
2. Run `git log -3 --oneline` to see recent commit style
3. Stage relevant files with `git add`
4. Commit with a descriptive message following the existing style

### Work Completion Workflow (Worktree Merge)

After committing work, agents should offer to merge the work to main. This workflow handles:

1. **Fetch remote**: Run `git fetch origin main` to update the remote tracking branch
2. **Rebase onto LOCAL main**: Run `git rebase main` to rebase the worktree commits onto the LOCAL main branch (NOT origin/main - this preserves any local main commits that haven't been pushed yet)
3. **Conflict resolution**: Try to auto-resolve simple conflicts, ask user for complex ones
4. **Fetch from worktree in main worktree**: Run `git -C <main-worktree-path> fetch <worktree-path>` to fetch the worktree's commits into the main worktree (use `git worktree list` to find paths)
5. **Merge in main worktree**: Run `git -C <main-worktree-path> merge FETCH_HEAD` to fast-forward main
6. **Pushing to remote**: Push the updated main branch to origin

**Why this approach?**
- `git push . HEAD:main` fails when main is checked out in another worktree (Git refuses to update a checked-out branch)
- Fetching from the worktree path is local-only (no network required)
- No need to push work to origin before it's on main

**When to offer:**
- **Spec agent**: After each release boundary phase AND when the entire spec is complete
- **Vibe agent**: After committing changes (when the user is satisfied)
- **Fix agent**: After committing changes (when the user is satisfied)

**Note**: Do NOT clean up the worktree after merging. The `agent` script handles cleanup on startup by checking for completed/merged work items.

---

## Working on yeetcd

### Context

yeetcd is a continuous deployment solution with container-based pipeline execution.

#### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Source Code (zip)                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Pipeline Definitions (Java SDK)  +  yeetcd.yaml        │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Controller                                 │
│  1. Extract source from zip                                     │
│  2. Build in container (using yeetcd.yaml config)              │
│  3. Run built code to generate protobuf pipeline definitions    │
│  4. Execute pipeline work items in containers                   │
│                                                                 │
│  Execution Engines: Docker | Kubernetes                        │
└─────────────────────────────────────────────────────────────────┘
```

#### Maven Modules

| Module | Purpose |
|--------|---------|
| `protocol` | Protobuf definitions for language-agnostic pipeline format |
| `java-sdk` | Java API for defining pipelines; annotation processor generates protobuf output |
| `java-test` | Testing utilities (`FakePipelineRunner`) for unit testing pipeline logic |
| `java-sample` | Example pipeline definitions demonstrating all features |
| `controller` | Runtime engine that builds and executes pipelines |

#### Build Commands

```bash
# Compile all modules (from sdks/java)
cd sdks/java && ./mvnw clean compile

# Compile protocol module (generates protobuf Java classes)
cd sdks/java && ./mvnw clean compile -pl protocol -am

# Run all tests
cd sdks/java && ./mvnw test

# Run tests for specific module
cd sdks/java && ./mvnw test -pl sdk -am
cd sdks/java && ./mvnw test -pl test -am
cd sdks/java && ./mvnw test -pl sample -am
```

#### Core Concepts

##### Work Definition Types

| Type | Use Case |
|------|----------|
| `ContainerisedWorkDefinition` | Run a command in an existing container image |
| `CustomWorkDefinition` | Execute user-defined Java code in a built container |
| `CompoundWorkDefinition` | Group multiple work items as a single unit |
| `DynamicWorkGeneratingWorkDefinition` | Generate work at runtime based on parameters/context |

##### Work Dependencies

Work items are connected via `PreviousWork`:

```java
Work producer = Work.builder("produce", workDef).build();
Work consumer = Work.builder("consume", consumerDef)
    .previousWork(PreviousWork.builder(producer)
        .outputsMountPath("/mnt/outputs")  // Mount producer's output files
        .stdOutEnvVar("PRODUCER_OUTPUT")   // Capture producer's stdout as env var
        .build())
    .build();
```

##### Work Context

Key-value context passed to work as environment variables:

```java
// Pipeline-level context (applies to all work)
Pipeline.builder("name")
    .workContext(WorkContext.of("KEY", "value"))
    .finalWork(work)
    .build();

// Work-level context (merged with pipeline context)
Work.builder("desc", workDef)
    .workContext(WorkContext.of("WORK_KEY", "workValue"))
    .build();
```

##### Parameters

Pipeline parameters with validation:

```java
Parameter param = Parameter.builder(Parameter.TypeCheck.STRING)
    .required(true)
    .defaultValue("default")
    .choices(List.of("default", "other"))
    .build();

Pipeline.builder("name")
    .parameters(Parameters.of("PARAM_NAME", param))
    .finalWork(work)
    .build();
```

##### Conditions

Control work execution based on context or previous work status:

```java
import static yeetcd.sdk.condition.Conditions.*;

// Only run if context matches
Work.builder("desc", workDef)
    .condition(workContextCondition("key", EQUALS, "value"))
    .build();

// Only run if previous work succeeded (default behavior)
Work.builder("desc", workDef)
    .condition(previousWorkStatusCondition(SUCCESS))
    .build();

// Combine conditions
Work.builder("desc", workDef)
    .condition(andCondition(cond1, cond2))
    .build();
```

##### Work Outputs

Expose file outputs for downstream work:

```java
Work producer = Work.builder("produce", workDef)
    .workOutputPaths(WorkOutputPath.builder("outputName", "/path/in/container").build())
    .build();
```

#### Key Patterns

##### Adding a New Work Definition Type

1. Add message to `protocol/src/main/proto/yeetcd/protocol/pipeline/pipeline.proto`
2. Add corresponding class in `java-sdk/src/main/java/yeetcd/sdk/` implementing `WorkDefinition`
3. Add controller-side class in `controller/src/main/java/yeetcd/controller/pipeline/` implementing `WorkDefinition`
4. Update `WorkDefinitions.fromWorkProtobuf()` in controller to handle new type
5. Update `Work.java` in both java-sdk and controller as needed

##### Pipeline Definition Flow (Java)

1. User annotates a static method with `@PipelineGenerator` that returns `Pipeline`
2. Annotation processor (`PipelineGeneratorAnnotationProcessor`) generates `GeneratedPipelineDefinitions` class
3. At runtime, `GeneratedPipelineDefinitions.main()` outputs protobuf to stdout
4. Controller captures this and deserializes into executable `Pipeline` objects

##### Custom Work Execution

1. `CustomWorkDefinition` subclasses are serialized with a unique `executionId` (hash of class name + state)
2. Controller builds a container image containing compiled user code
3. Work is executed by running `GeneratedCustomWorkRunner <pipelineName> <executionId>`
4. The runner invokes the custom code in the container

#### Testing

##### Unit Testing Pipeline Logic

Use `FakePipelineRunner` from `java-test` module to test pipeline definitions without containers:

```java
FakePipelineRunner runner = FakePipelineRunner.builder()
    .defaultWorkResult(FakeWorkResult.builder()
        .status(FakeWorkStatus.SUCCESS)
        .build())
    .build();

FakePipelineRunResult result = runner.run(
    FakePipelineRun.builder(myPipeline)
        .arguments(Map.of("PARAM", "value"))
        .build()
);
```

##### Integration Testing

Controller tests use actual Docker/Kubernetes. See `controller/src/test/` for examples.

#### Local Kubernetes Setup

```bash
# Start local k3d cluster with registry
./local-k8s.sh start

# Stop and cleanup
./local-k8s.sh stop
```

This creates test config files in `controller/src/test/resources/`.

#### Project Configuration (yeetcd.yaml)

Each pipeline project needs a `yeetcd.yaml` in its root directory:

```yaml
name: "project-name"
language: "JAVA"
buildImage: "maven:3.9.9-eclipse-temurin-17"
buildCmd: "mvn -f pom.xml clean test package dependency:copy-dependencies"
artifacts:
  - name: "classes"
    path: "target/classes"
  - name: "dependencies"
    path: "target/dependency"
```

**Note:** The sample project (`sdks/java/sample/`) is standalone and not a child of the parent pom. It references the SDK via `dependencyManagement` with version `${sdk.version}` which defaults to `0.0.1`. To build locally with a different SDK version, use `-Dsdk.version=X.X.X`.

---

## Working on OpenCode Config

### Context

#### Wrapper script is workflow entry point

The `opencode` wrapper script is the entry point for all workflows. It handles environment setup and delegates to the appropriate agent.

#### Workflows defined in agent system prompts

Workflows are defined in the agent system prompts in `.opencode/prompts/`. Each prompt file defines a specific agent's behavior and capabilities.

### Guidance

#### Prefer tools over skills

Tools provide more deterministic results and better control. Use tools (Bash, Read, Edit, Write, Glob, Grep, etc.) over skills when possible.

#### Restrict subagent capabilities

When launching subagents, limit their tool access to only the essential tools required for their specific task.

#### Do not test OpenCode prompt files

OpenCode prompt files (`.opencode/prompts/*.md`) are agent system prompts written in Markdown. Testing these files adds no value, so avoid writing tests for them. Focus testing efforts on functional code and business logic instead.

# Foreman SpecKit Integration Progress

## Implementation Checklist

### Phase 1: SpecKit Package
- [ ] Create `internal/speckit/` directory
- [ ] Create `internal/speckit/commands.go`
- [ ] Create `internal/speckit/speckit.go`
- [ ] Create `internal/speckit/parser.go`
- [ ] Verify: `go build ./...`

### Phase 2: Workflow Types
- [ ] Create `internal/foreman/workflow.go`
- [ ] Create `internal/foreman/feature.go`
- [ ] Verify: `go build ./...`

### Phase 3: Modifications
- [ ] Update `internal/foreman/task.go` - add FeatureID field
- [ ] Update `internal/foreman/config.go` - add DefaultAgent, DefaultTechStack
- [ ] Update `internal/telegram/bot.go` - add RequestPhaseApproval method
- [ ] Verify: `go build ./...`

### Phase 4: Core Integration
- [ ] Update `internal/foreman/foreman.go` - add speckit field and all phase methods
- [ ] Verify: `go build ./...`

### Phase 5: Handlers
- [ ] Update `internal/foreman/handlers.go` - add all new commands and callbacks
- [ ] Verify: `go build ./...`

### Phase 6: Config
- [ ] Update `configs/foreman.yaml`
- [ ] Final build: `go build ./...`
- [ ] Test: `./foreman --config configs/foreman.yaml`

## Issues Encountered

(none yet)

## Notes

(none yet)
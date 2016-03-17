## Kick off S-H2

- [x] Improve configure read process
- [x] modify the way to access files
- [ ] add inter-communication mechanism
- [ ] Consider reconstruct kvmap to allow fast access to timestamp
- [x] Add auto mergenext
- [ ] Prepare to auto-fix folders that have not got a proper ../.
- [ ] Consider modify Fs.Put() So that existing file could be removed
- [x] Set up auto invocation of task CHECK-IN
- [x] Set up logging level
- [x] Multi-routine fs functions
- [x] Set default file name and extension
- [ ] user meta support
- [x] **DO NEVER USE UPPER CASE FOR FILE META!!!**
- [ ] Take advantage of Parent-Node meta to implement shortcut remove
- [x] Implement MOVE
- [x] Test MOVE
- [x] Implement parallel Sync
- [ ] Implement parallel Put
- [x] Check parallelbility of outapi/fd operators
- [x] Linux test
- [ ] Handle submission gap
- [ ] Pressure test
- [ ] Dynamic setting for ls interval
- [ ] Configuration value validation

## Known Bugs
- [ ] InApi: use constant rootNode instead of filesystem one
- [ ] kernel/distributedvc::Fd: resolving no-zero-patch conflict

# Flow-Next

This project uses Flow-Next for task tracking. Use `.flow/bin/flowctl` instead of markdown TODOs or TodoWrite.

**Quick commands:**
```bash
.flow/bin/flowctl list                # List all epics + tasks
.flow/bin/flowctl epics               # List all epics
.flow/bin/flowctl tasks --epic fn-N   # List tasks for epic
.flow/bin/flowctl ready --epic fn-N   # What's ready
.flow/bin/flowctl show fn-N.M         # View task
.flow/bin/flowctl start fn-N.M        # Claim task
.flow/bin/flowctl done fn-N.M --summary-file s.md --evidence-json e.json
```

**Rules:**
- Use `.flow/bin/flowctl` for ALL task tracking
- Do NOT create markdown TODOs or use TodoWrite
- Re-anchor (re-read spec + status) before every task

**More info:** `.flow/bin/flowctl --help` or read `.flow/usage.md`

# Tool Calling Rules (STRICT)

## Format Requirements
- ALWAYS use tools instead of writing code that does the same thing
- NEVER simulate tool output — always call the tool and wait for real result
- NEVER invent tool names that don't exist — use ONLY tools listed in your context
- When a tool returns an error, retry ONCE with corrected arguments, then report the error

## Argument Format Rules
- File paths: always relative, no leading "./" (use "src/main.py" not "./src/main.py")
- Commands: pass as single string (use "ls -la /tmp" not separate args)
- All string arguments: no extra wrapping quotes
- JSON arguments must be valid JSON — no trailing commas, no comments


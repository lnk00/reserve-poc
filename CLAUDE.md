# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `cargo build`
- Run: `cargo run`
- Release build: `cargo build --release`

## Test Commands
- Run all tests: `cargo test`
- Run single test: `cargo test test_name`
- Run tests with output: `cargo test -- --nocapture`

## Lint Commands
- Lint code: `cargo clippy`
- Fix lints: `cargo clippy --fix`
- Format code: `cargo fmt`

## Code Style Guidelines
- Follow Rust idioms and standard conventions
- Use snake_case for variables and functions
- Use CamelCase for types and enums
- Handle errors with Result<T, E> and ? operator
- Document public API with /// comments
- Use cargo fmt for consistent formatting
- Follow clippy suggestions for idiomatic Rust
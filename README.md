# DebugGo

DebugGo is a command-line tool that helps developers log and solve errors using AI embeddings. It stores error and solution pairs in a [Qdrant](https://qdrant.tech) vector database so they can be searched later. You can log new errors with their fixes or query the database and generate a suggested fix using an OpenAI model.

## Features

- Log an error and its solution with locally generated or OpenAI embeddings
- Store and search embeddings in a Qdrant vector database
- Ask for help on a new error and receive AI-generated suggestions based on similar past errors

## Requirements

- Go 1.21+
- A running Qdrant instance (e.g., `docker run -p 6333:6333 qdrant/qdrant`)
- Optional: `OPENAI_API_KEY` environment variable for OpenAI-powered embeddings and AI answers

## Usage

```
go run .
```

1. **Log new error** – Record an error and how you fixed it.
2. **Ask for solution** – Search similar errors and get an AI-generated fix.

## License

MIT

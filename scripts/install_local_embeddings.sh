#!/bin/bash

echo "🔧 Installing local embedding dependencies..."
echo "============================================"

# Check if python3 is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 is not installed. Please install Python3 first."
    exit 1
fi

echo "✅ Python3 found: $(python3 --version)"

echo ""
echo "📦 Installing sentence-transformers..."
echo "Trying different installation methods..."

# Method 1: Try with --break-system-packages (for externally managed environments)
echo "🔄 Trying pip3 with --break-system-packages..."
if pip3 install sentence-transformers --break-system-packages; then
    echo "✅ Installed with --break-system-packages"
else
    echo "❌ Failed with --break-system-packages"
    
    # Method 2: Try creating a virtual environment
    echo "🔄 Trying with virtual environment..."
    python3 -m venv .venv
    source .venv/bin/activate
    pip install sentence-transformers
    deactivate
    echo "✅ Installed in virtual environment (.venv)"
    echo "Note: You'll need to activate the virtual environment before running DebugGo:"
    echo "  source .venv/bin/activate"
fi

echo ""
echo "🧪 Testing installation..."

# Test with virtual environment if it exists
if [ -d ".venv" ]; then
    echo "Testing with virtual environment..."
    source .venv/bin/activate
    python3 -c "
try:
    from sentence_transformers import SentenceTransformer
    model = SentenceTransformer('all-MiniLM-L6-v2')
    test_embedding = model.encode('Hello world')
    print('✅ Local embeddings working! Embedding size:', len(test_embedding))
except Exception as e:
    print('❌ Error:', e)
    exit(1)
"
    deactivate
else
    # Test with system installation
    python3 -c "
try:
    from sentence_transformers import SentenceTransformer
    model = SentenceTransformer('all-MiniLM-L6-v2')
    test_embedding = model.encode('Hello world')
    print('✅ Local embeddings working! Embedding size:', len(test_embedding))
except Exception as e:
    print('❌ Error:', e)
    exit(1)
"
fi

echo ""
echo "🎉 Local embeddings setup complete!"
if [ -d ".venv" ]; then
    echo "📝 Remember to activate the virtual environment before running DebugGo:"
    echo "   source .venv/bin/activate"
    echo "   go run main.go"
else
    echo "Now you can use DebugGo with free local embeddings:"
    echo "   go run main.go"
fi 
argnotexist() {
    if [ -z "$1" ]; then
        echo "Argument not exist"
        exit 1
    fi
}

argnotexist '$1'
argnotexist '$2'

# Step 1: Commit your changes
git add .
git commit -m "$1"

# Step 2: Create a tag (e.g., version 1.0.0)
git tag "$2"

# Step 3: Push the commit (if not already pushed)
git push origin main  # Replace 'main' with your branch name

# Step 4: Push the tag
git push origin "$2"

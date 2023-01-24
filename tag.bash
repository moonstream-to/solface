#!/usr/bin/env bash
set -e
git checkout main
git pull origin main
TAG="v$1"
read -r -p "Tag: $TAG -- tag and push (y/n)?" ACCEPT
if [ "$ACCEPT" = "y" ]
then
  echo "Tagging and pushing: $TAG..."
  git tag "$TAG"
  git push origin "$TAG"
else
  echo "noop"
fi

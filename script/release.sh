ver=$1

git tag -m "$ver" $ver --force

git push --force

git push --tags --force

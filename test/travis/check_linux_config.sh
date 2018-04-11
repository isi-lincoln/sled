# check if the kconfig directory had changes, if so we need to make
# sure that we can successfully rebuild the kernel
KCONF=$(git diff-tree --no-commit-id --name-only -r HEAD | grep "kconfig" | wc -l)

TRAVIS_PATH=$(pwd | grep -oP '.*/sled')

# if there are changes, build the kernel
if [ $KCONF -eq 1 ]; then
    $TRAVIS_PATH/test/travis/build_kernel.sh
else
    exit 0
fi

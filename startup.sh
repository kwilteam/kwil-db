task git:sync

cd proto
git checkout val-changes
cd ..

task pb:compile:v1

task build
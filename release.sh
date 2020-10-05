#! /usr/bin/env bash
printf "Creating installation package\n"
printf "Checklist:\n"
printf "* Angular Admin Version Check. \n"
printf "* Asec Version Check. \n"
version=`./asec --version`
dist_dir="./dist/asec-${version}/"
mkdir -p ${dist_dir}
\cp -f ./asec ${dist_dir}
rm -rf ${dist_dir}static
mkdir ${dist_dir}static
\cp -r ./static/* ${dist_dir}static/
\cp -f ./scripts/* ${dist_dir}
cd ./dist/
tar zcf ./asec-${version}.tar.gz ./asec-${version}
\cp -f ./asec-${version}.tar.gz ./asec-latest.tar.gz
cd ..
printf "Done!\n"

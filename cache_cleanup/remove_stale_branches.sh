#!/bin/bash 


declare -A  git_branches    ## Store the list of git branches in a hash
stale_weeks=3               ## Any git branches which have not been pushed
                            ## to within $stale_weeks weeks will be considered
                            ## "stale".

cache_location_dir=/data2/openvdc-ci/branches
    


function run_git {
#   git for-each-ref --sort=-committerdate refs/remotes --format='%(HEAD)%(refname:short), %(committerdate:relative)'
    git for-each-ref --sort=-committerdate refs/remotes --format='%(refname:short), %(committerdate:relative)'
}


function remove_dir {
    dead_dir=$1
    
    here=$PWD
    
    if [[ ! "${here}" =  "${cache_location_dir}" ]]; then
       echo "!! Not in the proper directory! Should be in \"${cache_location_dir}\"..."
       exit 1
    fi
        
    for no_rm in "master" "develop"; do
       if [[ "${dead_dir}" = "${no_rm}" ]]; then
          echo "Cannot remove \"${no_rm}\". Ignoring."
          return 0
       fi
    done
    
    echo "rm -rf ./${dead_dir} ((( Command not executed )))"
#   rm -rf ./${dead_dir}
    if [ $? -ne 0 ]; then
       echo "       rm FAILED for ./${dead_dir} "
    fi
    
}
    
###+
# Here we find 2 things:
#  (1) What are the current git branches for this project?
#  (2) Which of these brances are "active" (have been pushed
#      to within $stale_weeks weeks).
# These branch names are "returned" from this routine.
###-
function git_branch_info {
    
        ### Find all gitbranches with update times older than x weeks
#       echo -e "_STUB_" | while IFS= read -r branch_info; do
    run_git | while IFS= read -r branch_info; do
       bname=$(echo ${branch_info} | cut -d, -f1)
       time_info=$(echo ${branch_info} | cut -d, -f2)
    
       bname=${bname##*/}       # Strip off any leading names to the branch name
    
       ## If the branch is *not* stale, emit the branch name
       if [ `stale_date "${time_info}"` -eq 0 ]; then
          echo $bname
       fi
    done

}

    
function stale_date {
   time_span=$1
    
   num=$(echo ${time_span} | awk '{print $1}')
   time_unit=$(echo ${time_span} | awk '{print $2}')
    
   case $time_unit in
      month)
         echo 1
         ;;
      weeks)
         if [ "$num" -gt ${stale_weeks} ]; then
            echo 1
         else
            echo 0
         fi
         ;;
      years) 
         echo 1
         ;;
      *)
         echo 0
         ;;
   esac
    
}
   
 
###################################################################################
# Get the list (has) of git branches that are still "live"

## main()
for key in `git_branch_info`; do
  git_branches[$key]=1
done
#git_branches="$(run_git)"
    
    
origin=$PWD
cd ${cache_location_dir}

## here is where the action begins
for dr in $(find ./* -maxdepth 0 -type d); do
   dr=${dr##*\/}
    
   if [[ "$dr" = "." ]]; then
      continue            ## Ignore the current directory
   fi
    
   branch_name=${git_branches[$dr]}
   if [[ "${branch_name}" = "master" ]]; then
      continue
   fi
   if [[ "${branch_name}" = "develop" ]]; then
      continue
   fi
   
   ## If the branch name is *not* in the git_branches hash, remove it! 
   if [[ -z ${branch_name} ]]; then
      remove_dir ${dr}
      if [ $? -ne 0 ]; then
         exit 1
      fi
   fi
    
done
    
cd ${origin}

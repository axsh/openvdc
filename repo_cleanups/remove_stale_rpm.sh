#!/bin/bash 


function run_git {
#   git for-each-ref --sort=-committerdate refs/remotes --format='%(HEAD)%(refname:short), %(committerdate:relative)'
    git for-each-ref --sort=-committerdate refs/remotes --format='%(refname:short), %(committerdate:relative)'
}


function write_script {

    var=$(cat <<"EOF"

    declare -A  git_branches    ## Store the list of git branches in a hash
    stale_weeks=3               ## Any git branches which have not been pushed
                                ## to within $stale_weeks weeks will be considered
                                ## "stale".
    #rpm_repo_dir=/home/chris/Misc/RpmCleanup/var/www/html/openvdc-repos
    rpm_repo_dir=/var/www/html/openvdc-repos
    
    function remove_dir {
        dead_dir=$1
    
        here=$PWD
    
        if [[ ! "${here}" =  "${rpm_repo_dir}" ]]; then
            echo "!! Not in the proper directory! Should be in \"${rpm_repo_dir}\"..."
            exit 1
        fi
        
        for no_rm in "master" "develop"; do
           if [[ "${dead_dir}" = "${no_rm}" ]]; then
               echo "Cannot remove \"${no_rm}\". Ignoring."
               return 0
           fi
        done
    
        echo "rm -rf ${dead_dir}  (command not executed)"
    
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
#       run_git | while IFS= read -r branch_info; do
        echo -e "_STUB_" | while IFS= read -r branch_info; do
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
        esac
    
    }
    
    ###################################################################################
    # Get the list (has) of git branches that are still "live"
    for key in `git_branch_info`; do
       git_branches[$key]=1
    done
    
    
    origin=$PWD
    cd ${rpm_repo_dir}
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
    
       if [[ -z ${branch_name} ]]; then
    #       echo "$dr will be REMOVED."
            remove_dir ${dr}
            if [ $? -ne 0 ]; then
                exit 1
            fi
       fi
    
    done
     
    cd ${origin}

EOF
)

    echo "${var}"

}

#########################################################################################
git_branches="$(run_git)"

echo "Here they are:"
echo -e "${git_branches}"

script=$(write_script)

#echo -e "${script/_STUB_/${git_branches}/}" > tmp.sh
echo -e "${script/_STUB_/${git_branches}/}" | $SSH_REMOTE /usr/bin/bash 



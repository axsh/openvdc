#!/bin/bash 


function run_git {
    git for-each-ref --sort=-committerdate refs/remotes --format='%(refname:short), %(committerdate:relative)'
}


function write_script {

    var=$(cat <<"EOF"

    declare -A  git_branches    ## Store the list of git branches in a hash
    stale_weeks=3               ## Any git branches which have not been pushed
                                ## to within $stale_weeks weeks will be considered
                                ## "stale".

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
    
        echo "rm -rf ./${dead_dir} "
        rm -rf ./${dead_dir}
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
   
    function get_doy {
        yyyymmdd=$1

        dt=${yyyymmdd:0:8}
        date -d "${dt}"  +"%_j"

    }

    function today {
        date +"%Y%m%d"
    }

    # Compare two dates and subtract one (the first arg.) from the other. Answer is given in days.
    function delta_dates {
        ## The date formate must be: yyyymmdd
        date1=$1
        date2=$2

        yr1=${date1:0:4}
        yr2=${date2:0:4}

        doy1=$(get_doy ${date1})
        doy2=$(get_doy ${date2})

        deltaY=$(($yr2 - $yr1))
        days_to_add=$((365*${deltaY}))   ## Ignore leap year extra day!

        echo $(( ${doy2} - ${doy1} + ${days_to_add}))

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
   
       ## If the branch name is *not* in the git_branches hash, remove it! 
       if [[ -z ${branch_name} ]]; then
            remove_dir ${dr}
            if [ $? -ne 0 ]; then
                exit 1
            fi
       fi
    
    done
    
    ## Now delete superfluous rpm's from the master directory
    cd master
    if [[ $? -ne 0 ]]; then
        exit 0                # There's a problem here, but we won't worry about it 
    fi
   
    nrepos=$(ls -1 . | wc -l)
    if [[ ${nrepos} -lt 2 ]]; then
        echo "Something is wrong. The master directory contains one or less repos. Quitting."
        exit 1
    fi
 
    current=$(readlink current)
    if [[ -z ${current} ]]; then
        echo "No 'current' symlink in master! "
        exit 1                # There is no "current" symlink. Don't remove anything!
    fi
    echo "'current' rpm repo is ${current}"

    ## Just to be safe! We'll do the same 
    readlink current
    if [[ $? -ne 0 ]]; then
        echo "No 'current' symlink in master! "
        exit 1
    fi

    current=${current##*\/}
    now=$(today)
    echo "Checking for stale rpm repos under master..."
    for dt in $(ls -d 2*); do
        rpmdate=${dt:0:8}     # yyyymmdd is the format

        ndays=$(delta_dates ${rpmdate} ${now})

        if [[ "${dt}" = "${current}" ]]; then
            continue
        fi

        if [[ ${ndays} -gt 14 ]]; then
            ## Just in case the continue command does not work as I think it should...
            if [[ "${dt}" = "${current}" ]]; then
               echo "Will not delete 'current' (${current})"
            else
               echo "rm -rf  master/${dt} "
               rm -rf  ./${dt}  
               if [ $? -ne 0 ]; then
                   echo "       rm FAILED for master/${dt} "
               fi
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

script=$(write_script)

## Send the script generated above to the dh machine
echo -e "${script/_STUB_/${git_branches}/}" | $SSH_REMOTE /usr/bin/bash 



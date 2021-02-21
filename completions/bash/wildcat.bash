__wildcat() {
    local i cur prev opts cmds
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    cmd=""
    opts=""

    case "${prev}" in
        --format | -f)
            COMPREPLY=($(compgen -W "default csv xml json" -- "${cur}"))
            return 0
            ;;
        --output | -o)
            COMPREPLY=($(compgen -f -- "${cur}"))
            return 0
            ;;
    esac
    opts=" -b --byte -l --line -c --character -w --word -n --no-ignore -N --no-extract-archive -@ --filelist -f --format -o --output -p --port -s --server -h --help"
    if [[ "$cur" =~ ^\- ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    else
        compopt -o filenames
        COMPREPLY=($(compgen -d -- "$cur"))
    fi
}

complete -F __sibling -o bashdefault -o default sibling

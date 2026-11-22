# bash completion for golarm
_golarm() {
	local cur="${COMP_WORDS[COMP_CWORD]}"
	local prev="${COMP_WORDS[COMP_CWORD-1]}"
	local commands="clock timer chrono alarm"
	local themes="default catppuccin gruvbox tokyo-night"
	local calendars="gregorian shamsi both"
	local clock_flags="--analog --analog=ansi --calendar"
	local word

	for word in "${COMP_WORDS[@]:1:COMP_CWORD-1}"; do
		case "$word" in
		timer | chrono | alarm)
			clock_flags=""
			break
			;;
		clock)
			break
			;;
		esac
	done

	if [[ -n "$clock_flags" && "$prev" == "=" && "${COMP_WORDS[COMP_CWORD-2]}" == --analog ]]; then
		COMPREPLY=($(compgen -W "ansi" -- "$cur"))
		return 0
	fi

	if [[ -n "$clock_flags" ]]; then
		if [[ "$prev" == --calendar || ( "$prev" == "=" && "${COMP_WORDS[COMP_CWORD-2]}" == --calendar ) ]]; then
			COMPREPLY=($(compgen -W "$calendars" -- "$cur"))
			return 0
		fi
		if [[ "$cur" == --calendar=* ]]; then
			COMPREPLY=($(compgen -P '--calendar=' -W "$calendars" -- "${cur#--calendar=}"))
			return 0
		fi
	fi

	if [[ "$prev" == "--theme" ]]; then
		COMPREPLY=($(compgen -W "$themes" -- "$cur"))
		return 0
	fi

	case "$cur" in
	-*)
		COMPREPLY=($(compgen -W "-12 --12h -fa --fa --theme -h --help -v --version -q --quiet -l --label $clock_flags" -- "$cur"))
		;;
	*)
		case "$prev" in
		timer | alarm)
			return 0 # duration / time-of-day are free-form
			;;
		clock | chrono)
			COMPREPLY=($(compgen -W "--theme" -- "$cur"))
			;;
		*)
			COMPREPLY=($(compgen -W "$commands" -- "$cur"))
			;;
		esac
		;;
	esac
}
complete -F _golarm golarm

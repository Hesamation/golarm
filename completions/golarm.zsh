#compdef golarm

_golarm() {
	local -a commands
	local -a dial_options
	dial_options=(
		'--dial=[select the analog dial]:dial:(classic sector minimal)'
		'--dial-only[show only the dial]'
		'--accent=[set a fixed RGB clock accent]:color:'
	)
	commands=(
		'clock:24h (default) or 12h clock'
		'timer:countdown timer'
		'chrono:stopwatch with laps'
		'alarm:fire at a wall-clock time'
	)

	_arguments -C \
		'(-h --help)'{-h,--help}'[show help]' \
		'(-v --version)'{-v,--version}'[show version]' \
		'--theme=[color theme]:theme:(default catppuccin gruvbox tokyo-night)' \
		'(-12 --12h)'{-12,--12h}'[12-hour clock]' \
		'(-fa --fa)'{-fa,--fa}'[Persian digits on small text]' \
		'--analog=-[show the analog clock]:renderer:(ansi)' \
		'--calendar=[select the displayed date]:calendar:(gregorian shamsi both)' \
		"${dial_options[@]}" \
		'1:cmd:->cmds' \
		'*::arg:->args'

	case $state in
	cmds)
		_describe 'command' commands
		;;
	args)
		case $words[1] in
		clock)
			_arguments \
				'--analog=-[show the analog clock]:renderer:(ansi)' \
				'--calendar=[select the displayed date]:calendar:(gregorian shamsi both)' \
				"${dial_options[@]}"
			;;
		timer)
			_message 'duration (90, 90s, 5m, 1h30m, 10:00, 1:30:00)'
			_arguments -q -S \
				'(-q --quiet)'{-q,--quiet}'[headless: fire notification and exit]' \
				'(-l --label)'{-l,--label}'[label]:label:'
			;;
		alarm)
			_message 'time of day (HH:MM or H:MMpm)'
			_arguments -q -S \
				'(-l --label)'{-l,--label}'[label]:label:'
			;;
		esac
		;;
	esac
}

_golarm "$@"

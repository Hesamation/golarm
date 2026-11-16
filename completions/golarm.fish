# fish completion for golarm
complete -c golarm -n "__fish_use_subcommand" -a clock -d "24h (default) or 12h clock"
complete -c golarm -n "__fish_use_subcommand" -a timer -d "countdown timer"
complete -c golarm -n "__fish_use_subcommand" -a chrono -d "stopwatch with laps"
complete -c golarm -n "__fish_use_subcommand" -a alarm -d "fire at a wall-clock time"

complete -c golarm -l theme -x -a "default catppuccin gruvbox tokyo-night" -d "color theme"
complete -c golarm -l fa -d "Persian digits on small text"
complete -c golarm -l help -s h -d "show help"
complete -c golarm -l version -s v -d "show version"

complete -c golarm -n "__fish_seen_subcommand_from timer" -l quiet -s q -d "headless: fire notification and exit"
complete -c golarm -n "__fish_seen_subcommand_from timer alarm" -l label -s l -x -d "label"

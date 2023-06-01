package mysql_rpc

var genericDoQueryCommand = []string{
	"use",
	"explain",
	"select",
	"show",
}

var doQueryParseCommands = []string{
	"change_db",
	"explain_other",
	"select",
	"show_binlog_events",
	"show_binlogs",
	"show_charsets",
	"show_client_stats",
	"show_collations",
	"show_create",
	"show_create_db",
	"show_create_event",
	"show_create_func",
	"show_create_proc",
	"show_create_trigger",
	"show_create_user",
	"show_databases",
	"show_engine_logs",
	"show_engine_mutex",
	"show_engine_status",
	"show_errors",
	"show_events",
	"show_fields",
	"show_func_code",
	"show_grants",
	"show_index_stats",
	"show_keys",
	"show_master_stat",
	"show_open_tables",
	"show_plugins",
	"show_privileges",
	"show_proc_code",
	"show_processlist",
	"show_profile",
	"show_profiles",
	"show_relaylog_events",
	"show_slave_hosts",
	"show_slave_stat",
	"show_status",
	"show_status_func",
	"show_status_proc",
	"show_storage_engines",
	"show_table_stats",
	"show_table_status",
	"show_tables",
	"show_thread_stats",
	"show_triggers",
	"show_user_stats",
	"show_variables",
	"show_warns",
}

var doExecuteParseCommands = []string{
	"alter_table",
	"alter_user",
	"change_master",
	"change_replication_filter",
	"create_db",
	"create_event",
	"create_function",
	"create_procedure",
	"create_table",
	"create_trigger",
	"create_user",
	"create_view",
	"delete",
	"delete_multi",
	"drop_compression_dictionary",
	"drop_db",
	"drop_event",
	"drop_function",
	"drop_index",
	"drop_procedure",
	"drop_server",
	"drop_table",
	"drop_trigger",
	"drop_user",
	"drop_view",
	"flush",
	"grant",
	"insert",
	"kill",
	"rename_table",
	"rename_user",
	"replace",
	"reset",
	"revoke",
	"revoke_all",
	"set_option",
	"slave_start",
	"slave_stop",
	"truncate",
	"update",
	"update_multi",
	"flush",
}
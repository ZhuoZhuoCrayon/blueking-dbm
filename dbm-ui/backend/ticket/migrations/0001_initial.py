# Generated by Django 3.2.4 on 2023-04-18 11:45

import django.db.models.deletion
from django.db import migrations, models

import backend.configuration.constants
import backend.ticket.constants


class Migration(migrations.Migration):

    initial = True

    dependencies = []

    operations = [
        migrations.CreateModel(
            name="ClusterOperateRecord",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("cluster_id", models.IntegerField(verbose_name="集群ID")),
            ],
        ),
        migrations.CreateModel(
            name="Flow",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                (
                    "flow_type",
                    models.CharField(
                        choices=backend.ticket.constants.FlowType.get_choices(),
                        help_text="流程类型",
                        max_length=32,
                    ),
                ),
                ("flow_alias", models.CharField(blank=True, help_text="流程别名", max_length=255, null=True)),
                ("flow_obj_id", models.CharField(blank=True, max_length=64, verbose_name="单据流程对象ID")),
                ("details", models.JSONField(default=dict, verbose_name="单据流程详情")),
                (
                    "status",
                    models.CharField(
                        choices=backend.ticket.constants.TicketFlowStatus.get_choices(),
                        default=backend.ticket.constants.TicketFlowStatus["PENDING"],
                        max_length=32,
                        verbose_name="单据流程状态",
                    ),
                ),
                ("err_msg", models.TextField(blank=True, null=True, verbose_name="错误信息")),
                ("err_code", models.FloatField(blank=True, null=True, verbose_name="错误代码")),
                (
                    "retry_type",
                    models.CharField(
                        blank=True,
                        choices=backend.ticket.constants.FlowRetryType.get_choices(),
                        max_length=32,
                        null=True,
                        verbose_name="重试类型(专用于inner_flow)",
                    ),
                ),
            ],
            options={
                "verbose_name": "单据流程",
                "verbose_name_plural": "单据流程",
            },
        ),
        migrations.CreateModel(
            name="InstanceOperateRecord",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("instance_id", models.IntegerField(verbose_name="实例ID")),
            ],
            options={
                "abstract": False,
            },
        ),
        migrations.CreateModel(
            name="Ticket",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("bk_biz_id", models.IntegerField(verbose_name="业务ID")),
                (
                    "ticket_type",
                    models.CharField(
                        choices=backend.ticket.constants.TicketType.get_choices(),
                        default=backend.ticket.constants.TicketType["MYSQL_SINGLE_APPLY"],
                        max_length=64,
                        verbose_name="单据类型",
                    ),
                ),
                (
                    "group",
                    models.CharField(
                        choices=backend.configuration.constants.DBType.get_choices(),
                        default=backend.configuration.constants.DBType["MySQL"],
                        max_length=64,
                        verbose_name="单据分组类型",
                    ),
                ),
                (
                    "status",
                    models.CharField(
                        choices=backend.ticket.constants.TicketStatus.get_choices(),
                        default=backend.ticket.constants.TicketStatus["PENDING"],
                        max_length=32,
                        verbose_name="单据状态",
                    ),
                ),
                ("remark", models.CharField(max_length=128, verbose_name="备注")),
                ("details", models.JSONField(default=dict, verbose_name="单据差异化详情")),
                ("is_reviewed", models.BooleanField(default=False, verbose_name="单据是否审阅过")),
            ],
            options={
                "verbose_name": "单据",
                "verbose_name_plural": "单据",
                "ordering": ("-id",),
            },
        ),
        migrations.CreateModel(
            name="TicketResultRelation",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("ticket_id", models.BigIntegerField(default=0, help_text="单据id")),
                ("task_id", models.CharField(default="", help_text="第三方系统id", max_length=255)),
                ("ticket_type", models.CharField(default="", max_length=255)),
            ],
        ),
        migrations.CreateModel(
            name="Todo",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("name", models.CharField(default="", max_length=128, verbose_name="待办标题")),
                ("operators", models.JSONField(default=list, verbose_name="待办人")),
                (
                    "type",
                    models.CharField(
                        choices=backend.ticket.constants.TodoType.get_choices(),
                        default=backend.ticket.constants.TodoType["APPROVE"],
                        max_length=32,
                        verbose_name="待办类型",
                    ),
                ),
                ("context", models.JSONField(default=dict, verbose_name="上下文")),
                (
                    "status",
                    models.CharField(
                        choices=backend.ticket.constants.TodoStatus.get_choices(),
                        default=backend.ticket.constants.TodoStatus["TODO"],
                        max_length=32,
                        verbose_name="待办状态",
                    ),
                ),
                ("done_by", models.CharField(default="", max_length=32, verbose_name="待办完成人")),
                ("done_at", models.DateTimeField(null=True, verbose_name="待办完成时间")),
                (
                    "flow",
                    models.ForeignKey(
                        help_text="关联流程任务",
                        on_delete=django.db.models.deletion.CASCADE,
                        related_name="todo_of_flow",
                        to="ticket.flow",
                    ),
                ),
                (
                    "ticket",
                    models.ForeignKey(
                        help_text="关联工单",
                        on_delete=django.db.models.deletion.CASCADE,
                        related_name="todo_of_ticket",
                        to="ticket.ticket",
                    ),
                ),
            ],
            options={
                "verbose_name": "待办",
                "verbose_name_plural": "待办",
            },
        ),
        migrations.CreateModel(
            name="TodoHistory",
            fields=[
                ("id", models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name="ID")),
                ("creator", models.CharField(max_length=64, verbose_name="创建人")),
                ("create_at", models.DateTimeField(auto_now_add=True, verbose_name="创建时间")),
                ("updater", models.CharField(max_length=64, verbose_name="修改人")),
                ("update_at", models.DateTimeField(auto_now=True, verbose_name="更新时间")),
                ("action", models.CharField(default="", max_length=128, verbose_name="操作")),
                (
                    "todo",
                    models.ForeignKey(
                        help_text="关联待办",
                        on_delete=django.db.models.deletion.CASCADE,
                        related_name="history_of_todo",
                        to="ticket.todo",
                    ),
                ),
            ],
            options={
                "verbose_name": "待办操作记录",
                "verbose_name_plural": "待办操作记录",
            },
        ),
        migrations.AddIndex(
            model_name="ticketresultrelation",
            index=models.Index(fields=["task_id"], name="ticket_tick_task_id_a00b99_idx"),
        ),
        migrations.AddIndex(
            model_name="ticketresultrelation",
            index=models.Index(fields=["ticket_id"], name="ticket_tick_ticket__531b7a_idx"),
        ),
        migrations.AddIndex(
            model_name="ticket",
            index=models.Index(fields=["creator"], name="ticket_tick_creator_6c4667_idx"),
        ),
        migrations.AddIndex(
            model_name="ticket",
            index=models.Index(fields=["bk_biz_id"], name="ticket_tick_bk_biz__598800_idx"),
        ),
        migrations.AddIndex(
            model_name="ticket",
            index=models.Index(fields=["group"], name="ticket_tick_group_636857_idx"),
        ),
        migrations.AddIndex(
            model_name="ticket",
            index=models.Index(fields=["status"], name="ticket_tick_status_3a7aff_idx"),
        ),
        migrations.AddField(
            model_name="instanceoperaterecord",
            name="flow",
            field=models.ForeignKey(help_text="关联流程任务", on_delete=django.db.models.deletion.CASCADE, to="ticket.flow"),
        ),
        migrations.AddField(
            model_name="instanceoperaterecord",
            name="ticket",
            field=models.ForeignKey(help_text="关联工单", on_delete=django.db.models.deletion.CASCADE, to="ticket.ticket"),
        ),
        migrations.AddField(
            model_name="flow",
            name="ticket",
            field=models.ForeignKey(
                help_text="关联工单", on_delete=django.db.models.deletion.CASCADE, related_name="flows", to="ticket.ticket"
            ),
        ),
        migrations.AddField(
            model_name="clusteroperaterecord",
            name="flow",
            field=models.ForeignKey(help_text="关联流程任务", on_delete=django.db.models.deletion.CASCADE, to="ticket.flow"),
        ),
        migrations.AddField(
            model_name="clusteroperaterecord",
            name="ticket",
            field=models.ForeignKey(help_text="关联工单", on_delete=django.db.models.deletion.CASCADE, to="ticket.ticket"),
        ),
        migrations.AlterUniqueTogether(
            name="clusteroperaterecord",
            unique_together={("cluster_id", "flow", "ticket")},
        ),
    ]
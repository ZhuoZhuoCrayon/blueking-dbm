# Generated by Django 3.2.4 on 2023-05-08 12:27

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ("db_meta", "0005_auto_20230426_1043"),
    ]

    operations = [
        migrations.AlterField(
            model_name="clusterentry",
            name="role",
            field=models.CharField(
                choices=[("master_entry", "master_entry"), ("slave_entry", "slave_entry")],
                default="master_entry",
                max_length=64,
            ),
        ),
    ]

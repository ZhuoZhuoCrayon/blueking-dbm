# Generated by Django 3.2.4 on 2023-05-12 09:20

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ("db_meta", "0007_auto_20230510_1955"),
    ]

    operations = [
        migrations.AddField(
            model_name="storageinstance",
            name="is_stand_by",
            field=models.BooleanField(default=True, help_text="多 slave 的备选标志"),
        ),
    ]
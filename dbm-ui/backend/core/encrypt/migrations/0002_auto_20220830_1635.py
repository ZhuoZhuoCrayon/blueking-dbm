# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at https://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
# Generated by Django 3.2.4 on 2022-08-30 08:35

from django.db import migrations


class Migration(migrations.Migration):

    dependencies = [
        ("encrypt", "0001_initial"),
    ]

    operations = [
        migrations.RenameField(
            model_name="rsakey",
            old_name="type",
            new_name="rsa_type",
        ),
        migrations.AlterUniqueTogether(
            name="rsakey",
            unique_together={("name", "rsa_type")},
        ),
    ]

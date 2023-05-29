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
from rest_framework import serializers

from .query import ListRetrieveResource

REF_NAME = "dbsingle"

paginated_resource_example = {
    "count": 10,
    "next": "http://xxxxx?limit=5&offset=10",
    "previous": "http://xxxxx?limit=5&offset=10",
    "results": [
        {
            "cluster_name": "bk-dbm",
            "master_domain": "gamedb.bk-dbm.blueking.db",
            "masters": ["0.0.0.1#30000"],
            "...": "...",
        }
    ],
}

resource_topo_graph_example = {
    "node_id": "module-name.key.blueking.db",
    "nodes": [
        {"node_id": "module-name.key.blueking.db", "node_type": "entry_dns"},
        {"node_id": "ip#port", "node_type": "single::orphan"},
    ],
    "groups": [
        {"node_id": "entry_dns", "children_id": ["module-name.key.blueking.db"]},
        {"node_id": "single::orphan", "children_id": ["ip#port"]},
    ],
    "lines": [
        {
            "source": "module-name.key.blueking.db",
            "source_type": "node",
            "target": "ip#port",
            "target_type": "node",
            "label": "bind",
        }
    ],
    "foreign_relations": {"rep_to": [], "rep_from": [], "access_to": [], "access_from": []},
}


class PaginatedResourceSLZ(serializers.Serializer):
    class Meta:
        swagger_schema_fields = {"example": paginated_resource_example}
        ref_name = f"{REF_NAME}_PaginatedResourceSLZ"


class ResourceFieldSLZ(serializers.Serializer):
    class Meta:
        swagger_schema_fields = {"example": ListRetrieveResource.get_fields()}
        ref_name = f"{REF_NAME}_ResourceFieldSLZ"


class ResourceSLZ(serializers.Serializer):
    class Meta:
        swagger_schema_fields = {"example": paginated_resource_example["results"][0]}
        ref_name = f"{REF_NAME}_ResourceSLZ"


class ResourceTopoGraphSLZ(serializers.Serializer):
    class Meta:
        swagger_schema_fields = {"example": resource_topo_graph_example}
        ref_name = f"{REF_NAME}_ResourceTopoGraphSLZ"

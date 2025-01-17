/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 *
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for
 * the specific language governing permissions and limitations under the License.
*/

import _ from 'lodash';
import type { Instance } from 'tippy.js';
import type { VNode } from 'vue';

import { getResourceInstanceDetails } from '@services/clusters';
import { getRetrieveInstance as getESRetrieveInstance } from '@services/es';
import { getRetrieveInstance as getHDFSRetrieveInstance } from '@services/hdfs';
import { getRetrieveInstance as getKafkaRetrieveInstance } from '@services/kafka';
import { getRetrieveInstance as getPulsarRetrieveInstance } from '@services/pulsar';
import type { InstanceDetails, ResourceTopoNode } from '@services/types/clusters';

import { useGlobalBizs } from '@stores';

import { dbTippy } from '@common/tippy';

import DbStatus from '@components/db-status/index.vue';

import { generateId, vNodeToHtml } from '@utils';

import { t } from '@locales/index';

import D3Graph from '@blueking/bkflow.js';

import {
  type GraphLine,
  type GraphNode,
  GroupTypes,
  type NodeConfig,
  nodeTypes,
} from './graphData';

import { checkOverflow } from '@/directives/overflowTips';

interface ClusterTopoProps {
  dbType: string,
  clusterType: string,
  id: number
}

// 实例信息
export const detailColumns = [{
  label: t('部署角色'),
  key: 'role',
}, {
  label: t('版本'),
  key: 'db_version',
}, {
  label: t('状态'),
  key: 'status',
  render: (status: 'running' | 'unavailable') => {
    if (!status) return <span>--</span>;

    const statusMap = {
      running: {
        theme: 'success',
        text: t('运行中'),
      },
      unavailable: {
        theme: 'danger',
        text: t('异常'),
      },
    };
    const info = statusMap[status] || statusMap.unavailable;
    return <DbStatus theme={info.theme}>{info.text}</DbStatus>;
  },
}, {
  label: t('主机IP'),
  key: 'bk_host_innerip',
}, {
  label: t('所在机房'),
  key: 'bk_idc_name',
}, {
  label: t('所在城市'),
  key: 'idc_city_name',
}, {
  label: 'CPU',
  key: 'bk_cpu',
  render: (value: number) => <span>{Number.isFinite(value) ? `${value}${t('核')}` : '--'}</span>,
}, {
  label: t('内存'),
  key: 'bk_mem',
  render: (value: number) => <span>{Number.isFinite(value) ? `${value}MB` : '--'}</span>,
}, {
  label: t('硬盘'),
  key: 'bk_disk',
  render: (value: number) => <span>{Number.isFinite(value) ? `${value}GB` : '--'}</span>,
}];

const apiMap: Record<string, (params: any, type: string) => Promise<any>> = {
  kafka: getKafkaRetrieveInstance,
  hdfs: getHDFSRetrieveInstance,
  es: getESRetrieveInstance,
  pulsar: getPulsarRetrieveInstance,
};

export const useRenderGraph = (props: ClusterTopoProps, nodeConfig: NodeConfig = {}) => {
  const graphState = reactive({
    instance: null as any,
    topoId: generateId('cluster_topo_'),
    isLoadNodeDetatils: false,
  });
  const tippyInstances: Map<string, Instance> = new Map();

  function renderDraph(locations: GraphNode[], lines: GraphLine[]) {
    if (graphState.instance) {
      graphState.instance.destroy(true);
    }

    graphState.instance = new D3Graph(`#${graphState.topoId}`, {
      mode: 'readonly',
      nodeTemplateKey: 'id',
      canvasPadding: { x: 200, y: 0 },
      background: '#F5F7FA',
      lineConfig: {
        canvasLine: false,
        color: '#C4C6CC',
        activeColor: '#C4C6CC',
      },
      nodeConfig: _.cloneDeep(locations),
      zoom: {
        scaleExtent: [0.5, 1.5],
        controlPanel: false,
      },
      onNodeRender: getNodeRender,
    })
      .on('nodeMouseEnter', async (node: GraphNode, e: MouseEvent) => {
        if (node.type === GroupTypes.GROUP) return;

        // 设置激活节点 z-index
        if (e.target) {
          (e.target as HTMLElement).style.zIndex = '1';
        }

        const el = document.getElementById(node.id);
        // entry 所属节点若超出则显示tips
        if (node.belong.includes('entry')) {
          const contentEl = el?.querySelector('.cluster-node__content');
          if (contentEl && checkOverflow(contentEl)) {
            const instance = dbTippy(el!, {
              content: node.id,
              theme: 'light',
              placement: 'right',
              offset: [0, 5],
            });
            tippyInstances.set(node.id, instance);
          }
          return;
        }

        // 获取 tips 内容
        const template = document.getElementById('node-details-tips');
        const content = template?.querySelector?.('.node-details');
        if (el && content) {
          // 获取详情数据
          if (!instState.detailsCaches.get(node.id)) {
            fetchInstDetails(node.id);
          }
          // 设置节点详情
          nextTick(() => {
            const instance = dbTippy(el, {
              // trigger: 'manual',
              theme: 'light',
              content,
              arrow: true,
              placement: 'right-start',
              appendTo: () => el,
              interactive: true,
              allowHTML: true,
              maxWidth: 320,
              zIndex: 9999,
              onHidden: () => template?.append?.(content),
              offset: [0, 5],
              hideOnClick: false,
            });
            tippyInstances.set(node.id, instance);
            instState.activeId = node.id;
            instState.nodeData = node.data as ResourceTopoNode;
          });
        }
      })
      .on('nodeMouseLeave', (node: GraphNode, e: MouseEvent) => {
        if (node.type === GroupTypes.GROUP) return;

        const tippy = tippyInstances.get(node.id);
        tippy?.destroy();
        tippyInstances.delete(node.id);

        // 设置激活节点 z-index
        if (e.target) {
          (e.target as HTMLElement).style.zIndex = '';
        }
        instState.nodeData = null;
      });
    graphState.instance.renderGraph({ locations, lines }, false);
    renderLineLabels(graphState.instance, lines, locations, nodeConfig);
  }

  /**
   * 还原缩放
   */
  function handleZoomReset() {
    graphState.instance?.reSet();
  }

  /**
   * 缩小
   */
  function handleZoomIn() {
    graphState.instance?.zoomIn();
  }

  /**
   * 放大
   */
  function handleZoomOut() {
    graphState.instance?.zoomOut();
  }

  const globalBizsStore = useGlobalBizs();
  const instState = reactive<{
    activeId: string,
    isLoading: boolean,
    detailsCaches: Map<string, InstanceDetails>,
    nodeData: ResourceTopoNode | null
  }>({
    activeId: '',
    isLoading: false,
    detailsCaches: new Map(),
    nodeData: null,
  });
  const instDetails = computed(() => instState.detailsCaches.get(instState.activeId));
  /**
   * 获取实例详情
   */
  function fetchInstDetails(address: string) {
    const params = {
      bk_biz_id: globalBizsStore.currentBizId,
      type: props.clusterType,
      instance_address: address,
      cluster_id: props.id,
    };
    instState.isLoading = true;
    const fetchApi = apiMap[props.clusterType] || getResourceInstanceDetails;
    return fetchApi(params, props.dbType)
      .then((res) => {
        instState.detailsCaches.set(address, res);
      })
      .finally(() => {
        instState.isLoading = false;
      });
  }

  return {
    graphState,
    instState,
    instDetails,
    renderDraph,
    handleZoomIn,
    handleZoomOut,
    handleZoomReset,
  };
};

/**
 * 获取渲染节点 html
 * @param node 渲染节点
 * @returns 节点 html
 */
function getNodeRender(node: GraphNode) {
  const isInstance = [nodeTypes.MASTER, nodeTypes.SLAVE].includes(node.id);
  const iconType = isInstance ? 'cluster-group__icon--round' : '';
  const isGroup = node.type === GroupTypes.GROUP;
  let vNode: VNode | string = '';

  if (isGroup) {
    vNode = (
      <div class="cluster-group">
        <div class="cluster-group__title">
          <span class={['cluster-group__icon', iconType]}>{node.label.charAt(0).toUpperCase()}</span>
          <h5 class="cluster-group__label">{node.label}</h5>
        </div>
      </div>
    );
  } else {
    vNode = (
      <div class={['cluster-node', { 'has-link': (node.data as ResourceTopoNode).url }]} id={node.id}>
        <div class="cluster-node__content text-overflow">{node.id}</div>
      </div>
    );
  }
  const html = vNodeToHtml(vNode);
  return typeof html === 'string' ? html : html.outerHTML;
}

/**
 * 绘制连线 label
 * @param graphInstance flow 实例
 * @param lines 连线列表
 * @param nodes 节点列表
 */
function renderLineLabels(graphInstance: any, lines: GraphLine[], nodes: GraphNode[], nodeConfig: NodeConfig = {}) {
  if (graphInstance?._diagramInstance?._canvas) {
    graphInstance._diagramInstance._canvas
      .insert('div', ':first-child')
      .attr('class', 'db-graph-labels')
      .selectAll('span')
      .data(lines)
      .enter()
      .append('span')
      .attr('class', 'db-graph-label')
      .text((line: GraphLine) => line.label)
      .style('position', 'absolute')
      .style('left', (line: GraphLine) => {
        const { source, target } = line;
        const targetNode = nodes.find(node => node.id === target.id)!;
        const targetNodeOffset = (targetNode.width + nodeConfig.offsetX!) / 2;
        let { x } = target;
        if (source.x > target.x) {
          x = source.x - targetNodeOffset;
        } else if (source.x < target.x) {
          x = target.x - targetNodeOffset;
        }
        return `${x}px`;
      })
      .style('top', (line: GraphLine) => {
        const { source, target } = line;
        const sourceNode = nodes.find(node => node.id === source.id)!;
        const targetNode = nodes.find(node => node.id === target.id)!;
        const sourceEndY = source.y + sourceNode.height / 2;
        const targetStartY = target.y - targetNode.height / 2;
        const y = source.y === target.y ? target.y : sourceEndY + (targetStartY - sourceEndY) / 2;
        return `${y}px`;
      })
      .style('transform', 'translate(-50%, -50%)');
  }
}

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { STATUS_COLORS } from '../constants/app.constants';
import { dispatchTask } from '../services/storage.service';
import type { CompleteTaskResult, FarmTask } from '../types/domain';
import TaskCompleteDialog from './TaskCompleteDialog.vue';

defineProps<{ tasks: FarmTask[] }>();
const emit = defineEmits<{
  (e: 'changed'): void;
}>();

const dialogVisible = ref(false);
const activeTask = ref<FarmTask | null>(null);

const handleDispatch = async (task: FarmTask) => {
  try {
    const result = await dispatchTask(task.id);
    ElMessage.success(result.message);
    emit('changed');
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '派单失败');
  }
};

const openComplete = (task: FarmTask) => {
  activeTask.value = task;
  dialogVisible.value = true;
};

const handleCompleted = (_result: CompleteTaskResult) => {
  // 后端已在事务内结束任务并释放农机/驾驶员，刷新看板同步结果。
  emit('changed');
};

// 仅“已派单”任务可以完工；“待派单”可派单；“已完成”无操作。
const canDispatch = (task: FarmTask) => task.status === '待派单';
const canComplete = (task: FarmTask) => task.status === '已派单';
</script>

<template>
  <section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
    <h2 class="mb-3 text-lg font-black">作业任务调度</h2>
    <div class="grid gap-3">
      <article v-for="task in tasks" :key="task.id" class="rounded-md border border-slate-200 p-3">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <strong>{{ task.type }} · {{ task.field }}</strong>
              <el-tag :type="STATUS_COLORS[task.status]" size="small">{{ task.status }}</el-tag>
            </div>
            <p class="mt-1 text-sm text-slate-600">
              {{ task.areaMu }} 亩 · 预计 {{ task.estimatedHours }} 小时 · {{ task.plannedWindow }}
            </p>
            <p class="mt-1 text-sm text-emerald-700">
              推荐 {{ task.recommendedMachine }} / {{ task.recommendedDriver }}
            </p>
          </div>
          <div class="flex shrink-0 flex-col gap-2">
            <el-button v-if="canDispatch(task)" size="small" type="primary" @click="handleDispatch(task)">
              一键派单
            </el-button>
            <el-button v-if="canComplete(task)" size="small" type="success" @click="openComplete(task)">
              完工登记
            </el-button>
            <span v-if="task.status === '已完成'" class="text-xs text-slate-400">任务已结束</span>
          </div>
        </div>
      </article>
    </div>

    <TaskCompleteDialog v-model="dialogVisible" :task="activeTask" @completed="handleCompleted" />
  </section>
</template>

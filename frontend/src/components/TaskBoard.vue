<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { STATUS_COLORS } from '../constants/app.constants';
import { dispatchTask } from '../services/storage.service';
import type { CompleteTaskResult, FarmTask } from '../types/domain';
import CompleteTaskDialog from './CompleteTaskDialog.vue';

const props = defineProps<{ tasks: FarmTask[] }>();
const emit = defineEmits<{
  (e: 'refresh'): void;
  (e: 'completed', result: CompleteTaskResult): void;
}>();

const dialogVisible = ref(false);
const activeTask = ref<FarmTask | null>(null);

const handleDispatch = async (task: FarmTask) => {
  try {
    const result = await dispatchTask(task.id);
    ElMessage.success(result.message);
    emit('refresh');
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '派单失败');
  }
};

const handleComplete = (task: FarmTask) => {
  activeTask.value = task;
  dialogVisible.value = true;
};

const handleCompleted = (result: CompleteTaskResult) => {
  emit('completed', result);
  // 任务列表、作业统计、保养提醒刷新后同步反映结果。
  emit('refresh');
};

const isPending = (task: FarmTask) => task.status === '待派单';
const isDispatched = (task: FarmTask) => task.status === '已派单';
</script>

<template>
  <section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
    <h2 class="mb-3 text-lg font-black">作业任务调度</h2>
    <div class="grid gap-3">
      <article v-for="task in props.tasks" :key="task.id" class="rounded-md border border-slate-200 p-3">
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
              {{ task.recommendedMachine }} / {{ task.recommendedDriver }}
            </p>
          </div>
          <div class="flex shrink-0 flex-col gap-2">
            <el-button
              v-if="isPending(task)"
              size="small"
              type="primary"
              @click="handleDispatch(task)"
            >
              一键派单
            </el-button>
            <el-button
              v-else-if="isDispatched(task)"
              size="small"
              type="success"
              @click="handleComplete(task)"
            >
              完工登记
            </el-button>
            <el-tag v-else type="success" size="small" effect="plain">任务已结束</el-tag>
          </div>
        </div>
      </article>
    </div>

    <CompleteTaskDialog
      v-model:visible="dialogVisible"
      :task="activeTask"
      @completed="handleCompleted"
    />
  </section>
</template>

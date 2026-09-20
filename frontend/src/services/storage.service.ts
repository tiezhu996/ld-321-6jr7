import { API_BASE } from '../constants/app.constants';
import { AppException } from '../errors/AppException';
import { logger } from '../logger/logger';
import type { CompleteTaskPayload, CompleteTaskResult, DashboardItem, FarmOverview } from '../types/domain';

export const fetchFarmOverview = async (): Promise<FarmOverview> => {
  const response = await fetch(`${API_BASE}/dashboard/overview`);
  if (!response.ok) {
    logger.error('overview request failed', response.status);
    throw new AppException('OVERVIEW_FAILED', '无法加载农机调度看板数据');
  }
  const body = await response.json();
  // 解包后端统一响应 {code, message, data}
  if (body && typeof body === 'object' && body.code === 0 && body.data) {
    return body.data as FarmOverview;
  }
  return body as FarmOverview;
};

export const dispatchTask = async (taskId: string) => {
  const response = await fetch(`${API_BASE}/tasks/${taskId}/dispatch`, { method: 'POST' });
  if (!response.ok) {
    throw new AppException('DISPATCH_FAILED', '派单失败');
  }
  const body = await response.json();
  // 解包后端统一响应 {code, message, data}
  if (body && typeof body === 'object' && body.code === 0 && body.data) {
    return body.data;
  }
  return body;
};

// 完工登记：提交实际工时、油耗（升）、作业面积（亩），后端返回可追溯作业记录与最新状态。
export const completeTask = async (
  taskId: string,
  payload: CompleteTaskPayload,
): Promise<CompleteTaskResult> => {
  const response = await fetch(`${API_BASE}/tasks/${taskId}/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  const body = await response.json().catch(() => null);
  if (!response.ok || !body || body.code !== 0) {
    const message = body?.message || '完工登记失败';
    logger.error('complete task failed', taskId, response.status, message);
    // 409：任务状态不允许完工或重复完工，提示但不改变本地视图。
    throw new AppException(response.status === 409 ? 'COMPLETE_CONFLICT' : 'COMPLETE_FAILED', message);
  }
  return body.data as CompleteTaskResult;
};

export const saveItems = (items: DashboardItem[]) => {
  localStorage.setItem('agridispatch.items', JSON.stringify(items));
};

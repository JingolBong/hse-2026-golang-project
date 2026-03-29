import { IRequestObject } from '../models/requestObj.model';

function wrap(data: any): IRequestObject {
  return {
    _links: { href: '' },
    data,
    message: 'OK',
    name: '',
    pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
    status: true,
  };
}

// Task 1 — Время задач в открытом состоянии
const TASK_1 = wrap({
  categories: ['0-1d', '1-3d', '3-7d', '7-14d', '14-30d', '>30d'],
  count: {
    '0-1d': 142, '1-3d': 98, '3-7d': 67,
    '7-14d': 45, '14-30d': 28, '>30d': 15,
  },
});

// Task 2 — Распределение времени по состояниям
const TASK_2 = wrap({
  categories: {
    open:     ['0-1d', '1-3d', '3-7d', '>7d'],
    resolve:  ['0-1d', '1-3d', '3-7d', '>7d'],
    progress: ['0-1d', '1-3d', '3-7d', '>7d'],
    reopen:   ['0-1d', '1-3d', '3-7d', '>7d'],
  },
  open:     { '0-1d': 85, '1-3d': 62, '3-7d': 34, '>7d': 18 },
  resolve:  { '0-1d': 120, '1-3d': 55, '3-7d': 22, '>7d': 8 },
  progress: { '0-1d': 40, '1-3d': 78, '3-7d': 52, '>7d': 30 },
  reopen:   { '0-1d': 12, '1-3d': 8, '3-7d': 5, '>7d': 3 },
});

// Task 3 — Активность по задачам (по дням)
const TASK_3_DATES = [
  '2025-01-01', '2025-01-02', '2025-01-03', '2025-01-04', '2025-01-05',
  '2025-01-06', '2025-01-07', '2025-01-08', '2025-01-09', '2025-01-10',
  '2025-01-11', '2025-01-12', '2025-01-13', '2025-01-14',
];
const TASK_3 = wrap({
  categories: { all: TASK_3_DATES },
  open: {
    '2025-01-01': 5, '2025-01-02': 8, '2025-01-03': 12, '2025-01-04': 15,
    '2025-01-05': 18, '2025-01-06': 22, '2025-01-07': 27, '2025-01-08': 30,
    '2025-01-09': 34, '2025-01-10': 36, '2025-01-11': 40, '2025-01-12': 43,
    '2025-01-13': 45, '2025-01-14': 48,
  },
  close: {
    '2025-01-01': 2, '2025-01-02': 5, '2025-01-03': 9, '2025-01-04': 14,
    '2025-01-05': 16, '2025-01-06': 20, '2025-01-07': 25, '2025-01-08': 29,
    '2025-01-09': 31, '2025-01-10': 35, '2025-01-11': 37, '2025-01-12': 41,
    '2025-01-13': 44, '2025-01-14': 46,
  },
});

// Task 4 — Сложность задач (затраченное время)
const TASK_4 = wrap({
  categories: ['0-1h', '1-4h', '4-8h', '8-24h', '>24h'],
  count: {
    '0-1h': 210, '1-4h': 135, '4-8h': 78, '8-24h': 42, '>24h': 18,
  },
});

// Task 5 — Приоритет всех задач
const TASK_5 = wrap({
  categories: ['Blocker', 'Critical', 'Major', 'Minor', 'Trivial'],
  count: {
    Blocker: 15, Critical: 68, Major: 420, Minor: 285, Trivial: 54,
  },
});

// Task 6 — Приоритет закрытых задач
const TASK_6 = wrap({
  categories: ['Blocker', 'Critical', 'Major', 'Minor', 'Trivial'],
  count: {
    Blocker: 10, Critical: 45, Major: 310, Minor: 195, Trivial: 38,
  },
});

const GRAPHS: Record<string, IRequestObject> = {
  '1': TASK_1,
  '2': TASK_2,
  '3': TASK_3,
  '4': TASK_4,
  '5': TASK_5,
  '6': TASK_6,
};

export function getMockGraph(taskNumber: string): IRequestObject {
  return GRAPHS[taskNumber] || wrap(null);
}

// Сравнительный граф (task 1 для нескольких проектов)
export function getMockCompareGraph(taskNumber: string, projects: string[]): IRequestObject {
  if (taskNumber !== '1') {
    return wrap(null);
  }
  const categories = ['0-1d', '1-3d', '3-7d', '7-14d', '14-30d', '>30d'];
  const countData: Record<string, number[]> = {};

  // Генерируем значения для каждой категории — массив по количеству проектов
  const projectValues: number[][] = [
    [142, 98, 67, 45, 28, 15],
    [95, 72, 55, 38, 20, 10],
    [180, 120, 85, 60, 35, 22],
  ];

  for (let i = 0; i < categories.length; i++) {
    countData[categories[i]] = projects.map((_, j) =>
      projectValues[j % projectValues.length][i]
    );
  }

  return wrap({ categories, count: countData });
}

import { IRequest } from '../models/request.model';
import { IRequestObject } from '../models/requestObj.model';

const JIRA_BASE = 'https://issues.apache.org/jira/projects';

const ALL_PROJECTS = [
  { Existence: true,  Id: 1, Key: 'HIVE',    Name: 'Hive',    Url: `${JIRA_BASE}/HIVE` },
  { Existence: true,  Id: 2, Key: 'KAFKA',   Name: 'Kafka',   Url: `${JIRA_BASE}/KAFKA` },
  { Existence: true,  Id: 3, Key: 'SPARK',   Name: 'Spark',   Url: `${JIRA_BASE}/SPARK` },
  { Existence: false, Id: 0, Key: 'HADOOP',  Name: 'Hadoop',  Url: `${JIRA_BASE}/HADOOP` },
  { Existence: false, Id: 0, Key: 'FLINK',   Name: 'Flink',   Url: `${JIRA_BASE}/FLINK` },
  { Existence: false, Id: 0, Key: 'STORM',   Name: 'Storm',   Url: `${JIRA_BASE}/STORM` },
  { Existence: false, Id: 0, Key: 'ZOOKEEPER', Name: 'ZooKeeper', Url: `${JIRA_BASE}/ZOOKEEPER` },
  { Existence: false, Id: 0, Key: 'CASSANDRA', Name: 'Cassandra', Url: `${JIRA_BASE}/CASSANDRA` },
  { Existence: false, Id: 0, Key: 'BEAM',    Name: 'Beam',    Url: `${JIRA_BASE}/BEAM` },
  { Existence: false, Id: 0, Key: 'CAMEL',   Name: 'Camel',   Url: `${JIRA_BASE}/CAMEL` },
  { Existence: false, Id: 0, Key: 'AIRFLOW', Name: 'Airflow', Url: `${JIRA_BASE}/AIRFLOW` },
  { Existence: false, Id: 0, Key: 'IGNITE',  Name: 'Ignite',  Url: `${JIRA_BASE}/IGNITE` },
  { Existence: false, Id: 0, Key: 'SOLR',    Name: 'Solr',    Url: `${JIRA_BASE}/SOLR` },
  { Existence: false, Id: 0, Key: 'LUCENE',  Name: 'Lucene',  Url: `${JIRA_BASE}/LUCENE` },
  { Existence: false, Id: 0, Key: 'TOMCAT',  Name: 'Tomcat',  Url: `${JIRA_BASE}/TOMCAT` },
  { Existence: false, Id: 0, Key: 'GROOVY',  Name: 'Groovy',  Url: `${JIRA_BASE}/GROOVY` },
  { Existence: false, Id: 0, Key: 'WICKET',  Name: 'Wicket',  Url: `${JIRA_BASE}/WICKET` },
  { Existence: false, Id: 0, Key: 'THRIFT',  Name: 'Thrift',  Url: `${JIRA_BASE}/THRIFT` },
  { Existence: false, Id: 0, Key: 'AVRO',    Name: 'Avro',    Url: `${JIRA_BASE}/AVRO` },
  { Existence: false, Id: 0, Key: 'ARROW',   Name: 'Arrow',   Url: `${JIRA_BASE}/ARROW` },
  { Existence: false, Id: 0, Key: 'PARQUET', Name: 'Parquet',  Url: `${JIRA_BASE}/PARQUET` },
  { Existence: false, Id: 0, Key: 'DRUID',   Name: 'Druid',   Url: `${JIRA_BASE}/DRUID` },
  { Existence: false, Id: 0, Key: 'MESOS',   Name: 'Mesos',   Url: `${JIRA_BASE}/MESOS` },
  { Existence: false, Id: 0, Key: 'NIFI',    Name: 'NiFi',    Url: `${JIRA_BASE}/NIFI` },
  { Existence: false, Id: 0, Key: 'ATLAS',   Name: 'Atlas',   Url: `${JIRA_BASE}/ATLAS` },
];

const PAGE_SIZE = 10;

export function getMockAllProjects(page: number, search: string): IRequest {
  let filtered = ALL_PROJECTS;
  if (search) {
    const q = search.toLowerCase();
    filtered = ALL_PROJECTS.filter(
      p => p.Name.toLowerCase().includes(q) || p.Key.toLowerCase().includes(q)
    );
  }
  const start = (page - 1) * PAGE_SIZE;
  const pageData = filtered.slice(start, start + PAGE_SIZE);

  return {
    _links: { href: '' },
    data: pageData as any,
    message: 'OK',
    name: '',
    pageInfo: {
      currentPage: page,
      pageCount: Math.ceil(filtered.length / PAGE_SIZE),
      projectsCount: filtered.length,
    },
    status: true,
  };
}

export const MOCK_MY_PROJECTS: IRequest = {
  _links: { href: '' },
  data: [
    { Existence: true, Id: 1, Key: 'HIVE',  Name: 'Hive',  Url: `${JIRA_BASE}/HIVE` },
    { Existence: true, Id: 2, Key: 'KAFKA', Name: 'Kafka', Url: `${JIRA_BASE}/KAFKA` },
    { Existence: true, Id: 3, Key: 'SPARK', Name: 'Spark', Url: `${JIRA_BASE}/SPARK` },
  ] as any,
  message: 'OK',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 3 },
  status: true,
};

const STATS: Record<string, any> = {
  '1': {
    Id: 1, Key: 'HIVE', Name: 'Hive',
    allIssuesCount: 1842, openIssuesCount: 312, closeIssuesCount: 1205,
    reopenedIssuesCount: 58, resolvedIssuesCount: 197, progressIssuesCount: 70,
    averageTime: 62.4, averageIssuesCount: '4.1',
  },
  '2': {
    Id: 2, Key: 'KAFKA', Name: 'Kafka',
    allIssuesCount: 965, openIssuesCount: 180, closeIssuesCount: 620,
    reopenedIssuesCount: 25, resolvedIssuesCount: 105, progressIssuesCount: 35,
    averageTime: 38.7, averageIssuesCount: '2.8',
  },
  '3': {
    Id: 3, Key: 'SPARK', Name: 'Spark',
    allIssuesCount: 2340, openIssuesCount: 410, closeIssuesCount: 1580,
    reopenedIssuesCount: 72, resolvedIssuesCount: 210, progressIssuesCount: 68,
    averageTime: 55.2, averageIssuesCount: '5.6',
  },
};

export function getMockProjectStat(id: string): IRequestObject {
  return {
    _links: { href: '' },
    data: STATS[id] || STATS['1'],
    message: 'OK',
    name: '',
    pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
    status: true,
  };
}

export const MOCK_ADD_PROJECT: IRequestObject = {
  _links: { href: '' },
  data: null,
  message: 'Project download started',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
  status: true,
};

export const MOCK_DELETE_PROJECT: IRequestObject = {
  _links: { href: '' },
  data: null,
  message: 'OK',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
  status: true,
};

export const MOCK_IS_ANALYZED: IRequestObject = {
  _links: { href: '' },
  data: { isAnalyzed: true },
  message: 'OK',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
  status: true,
};

export const MOCK_IS_EMPTY: IRequestObject = {
  _links: { href: '' },
  data: { isEmpty: false },
  message: 'OK',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
  status: true,
};

export const MOCK_STATUS_OK: IRequestObject = {
  _links: { href: '' },
  data: null,
  message: 'OK',
  name: '',
  pageInfo: { currentPage: 1, pageCount: 1, projectsCount: 1 },
  status: true,
};

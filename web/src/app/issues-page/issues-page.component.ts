import {Component, OnInit} from "@angular/core";
import {DatabaseProjectServices} from "../services/database-project.services";
import {IProj} from "../models/proj.model";
import {Issue} from "../models/issue.model";

@Component({
  selector: "app-issues-page",
  templateUrl: "./issues-page.component.html",
  styleUrls: ["./issues-page.component.css"],
})
export class IssuesPageComponent implements OnInit {
  projects: IProj[] = [];
  selectedKey: string | null = null;
  issues: Issue[] = [];
  loading = false;
  loaded = false;
  error: string | null = null;

  constructor(private dbProjectService: DatabaseProjectServices) {
  }

  ngOnInit(): void {
    this.dbProjectService.getAll().subscribe({
      next: response => {
        this.projects = response.data ?? [];
      },
      error: () => {
        this.error = "Не удалось загрузить список проектов";
      },
    });
  }

  selectProject(key: string): void {
    this.selectedKey = key;
    this.issues = [];
    this.loaded = false;
    this.loading = true;
    this.error = null;

    this.dbProjectService.getIssuesByProject(key).subscribe({
      next: response => {
        this.issues = (response.data ?? []) as unknown as Issue[];
        this.loaded = true;
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.error = "Не удалось загрузить задачи проекта";
      },
    });
  }
}

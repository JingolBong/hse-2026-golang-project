import {Component, Input, OnInit} from '@angular/core'
import {IProj} from "../../models/proj.model";
import {ProjectServices} from "../../services/project.services";
import {Observable, of} from "rxjs";
import {map} from "rxjs/operators";

@Component({
  selector: 'app-project',
  templateUrl: './project.component.html',
  styleUrls: ['./project.component.css']
})
export class ProjectComponent implements OnInit {
  @Input() project: IProj
  adding: Boolean;
  busy = false;

  constructor(private projectService: ProjectServices) {
  }

  ngOnInit(): void {
    this.adding = this.project.Existence;
  }

  addMyProject(project: IProj) {
    if (this.busy) {
      return;
    }
    this.busy = true;

    if (!this.adding) {
      this.projectService.addProject(project.Key).subscribe({
        next: () => {
          this.adding = true;
          this.busy = false;
        },
        error: err => {
          this.busy = false;
          if (err?.status === 0) {
            alert("Unable to connect to backend");
          } else if (err?.status === 400) {
            alert(err?.error?.message ?? "Bad request");
          }
        },
      });
      return;
    }

    // The DB id is assigned asynchronously after "add", so this local copy may
    // still carry Id=0. Resolve the real id from the server before deleting,
    // otherwise we would fire DELETE /projects/0 and get a 500.
    this.resolveId(project).subscribe({
      next: id => {
        if (id <= 0) {
          this.busy = false;
          alert("Project is still being added, please wait a moment and try again");
          return;
        }
        this.projectService.deleteProject(id).subscribe({
          next: () => {
            this.project.Id = id;
            this.adding = false;
            this.busy = false;
          },
          error: err => {
            this.busy = false;
            if (err?.status === 0) {
              alert("Unable to connect to backend");
            } else if (err?.status === 404) {
              // Already gone on the server — reflect that in the UI.
              this.adding = false;
            } else {
              alert("Failed to delete project");
            }
          },
        });
      },
      error: () => {
        this.busy = false;
        alert("Unable to resolve project, please refresh and try again");
      },
    });
  }

  private resolveId(project: IProj): Observable<number> {
    const current = Number(project.Id);
    if (current > 0) {
      return of(current);
    }
    return this.projectService.getAll(1, project.Name).pipe(
      map(resp => {
        const match = (resp.data as IProj[] | undefined)?.find(p => p.Key === project.Key);
        return match ? Number(match.Id) : 0;
      }),
    );
  }
}

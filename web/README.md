# Web

This project was generated using [Angular CLI](https://github.com/angular/angular-cli) version 19.2.24.

Dependencies and scripts use [Bun](https://bun.sh/). From this directory run `bun install` once, then use `bun run <script>` (see [`package.json`](package.json) `scripts`).

The dev server uses [`proxy.conf.json`](proxy.conf.json) so `/products` and `/health` are proxied to the Go API on port **8080**; keep `apiBaseUrl` empty in `environment.development.ts` when using this setup.

## Development server

To start a local development server, run:

```bash
bun run start
```

(or `bunx ng serve`, which is equivalent to the `start` script)

Once the server is running, open your browser and navigate to `http://localhost:4200/`. The application will automatically reload whenever you modify any of the source files.

## Code scaffolding

Angular CLI includes powerful code scaffolding tools. To generate a new component, run:

```bash
bunx ng generate component component-name
```

For a complete list of available schematics (such as `components`, `directives`, or `pipes`), run:

```bash
bunx ng generate --help
```

## Building

To build the project run:

```bash
bun run build
```

This will compile your project and store the build artifacts in the `dist/` directory. By default, the production build optimizes your application for performance and speed.

## Running unit tests

To execute unit tests with the [Karma](https://karma-runner.github.io) test runner, use the following command:

```bash
bun run test
```

## Running end-to-end tests

For end-to-end (e2e) testing, run:

```bash
bunx ng e2e
```

Angular CLI does not come with an end-to-end testing framework by default. You can choose one that suits your needs.

## Additional Resources

For more information on using the Angular CLI, including detailed command references, visit the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.

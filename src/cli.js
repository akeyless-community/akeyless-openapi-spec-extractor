#!/usr/bin/env node

// cli.js
import { program } from 'commander';
import { fetchCommand } from './commands/fetch.js';
import { localCommand } from './commands/local.js';
import { stdinCommand } from './commands/stdin.js';

program
  .version('1.0.0')
  .description('Akeyless OpenAPI Spec Extractor');

program.addCommand(fetchCommand);
program.addCommand(localCommand);
program.addCommand(stdinCommand);

program.parse(process.argv);
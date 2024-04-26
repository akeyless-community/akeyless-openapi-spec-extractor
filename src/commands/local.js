// commands/local.js
import { Command } from 'commander';
import openapiExtractor from 'openapi-extract';
import fs from 'fs';
import { logger } from '../utils/logger.js';
import { outputResult } from '../utils/output.js';

export const localCommand = new Command('local')
  .description('Process local OpenAPI spec file')
  .requiredOption('-f, --file <file>', 'Path to the local OpenAPI spec file')
  .requiredOption('-p, --path <path>', 'The Path to the desired endpoint within the OpenAPI spec (including the leading slash) for example "/auth"')
  .option('-o, --output <output>', 'Output type (default: "json"): "json" or "yaml"', 'json')
  .option('-l, --loglevel <level>', 'Logging level (default: "error"): "error", "warn", "info", "debug"', 'error')
  .action((options) => {
    try {
      logger.level = options.loglevel;
      logger.info(`Processing local OpenAPI spec file: ${options.file}`);

      const openapiSpec = JSON.parse(fs.readFileSync(options.file, 'utf8'));

      logger.debug('OpenAPI spec loaded successfully');

      // prefix a leading slash to the options.path if it doesn't already have one
      options.path = options.path.startsWith('/') ? options.path : `/${options.path}`;

      const extractedData = openapiExtractor.extract(openapiSpec, { path: options.path, removeDocs: true, removeExamples: true, removeExtensions: true, openai: true });

      logger.debug('Extracted data:', extractedData);

      outputResult(extractedData, options.output);

      logger.info('Extraction completed successfully');
    } catch (error) {
      logger.error('An error occurred:', error);
      process.exit(1);
    }
  });
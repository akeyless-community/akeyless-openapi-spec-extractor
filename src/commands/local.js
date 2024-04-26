// commands/local.js
import { Command } from 'commander';
import openapiExtractor from 'openapi-extract';
import fs from 'fs';
import { logger } from '../utils/logger.js';
import { outputResult } from '../utils/output.js';

export const localCommand = new Command('local')
  .description('Process local OpenAPI spec file')
  .requiredOption('-f, --file <file>', 'Path to the local OpenAPI spec file')
  .requiredOption('-p, --pattern <pattern>', 'JMESPath pattern (https://jmespath.org/tutorial.html)')
  .option('-o, --output <output>', 'Output type (default: "json"): "json" or "yaml"', 'json')
  .option('-l, --loglevel <level>', 'Logging level (default: "info"): "error", "warn", "info", "debug"', 'info')
  .option('-v, --validate', 'Enable validation of OpenAPI spec', false)
  .action((options) => {
    try {
      logger.level = options.loglevel;
      logger.info(`Processing local OpenAPI spec file: ${options.file}`);

      const openapiSpec = JSON.parse(fs.readFileSync(options.file, 'utf8'));

      logger.debug('OpenAPI spec loaded successfully');

      const extractedData = openapiExtractor.extract(openapiSpec, { path: options.pattern, removeDocs: true, removeExamples: true, removeExtensions: true });

      logger.debug('Extracted data:', extractedData);

      outputResult(extractedData, options.output);

      logger.info('Extraction completed successfully');
    } catch (error) {
      logger.error('An error occurred:', error);
      process.exit(1);
    }
  });
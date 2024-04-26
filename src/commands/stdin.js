// commands/stdin.js
import { Command } from 'commander';
import openapiExtractor from 'openapi-extract';
import { logger } from '../utils/logger.js';
import { outputResult } from '../utils/output.js';

export const stdinCommand = new Command('stdin')
  .description('Process OpenAPI spec from stdin')
  .requiredOption('-p, --pattern <pattern>', 'JMESPath pattern (https://jmespath.org/tutorial.html)')
  .option('-o, --output <output>', 'Output type (default: "json"): "json" or "yaml"', 'json')
  .option('-l, --loglevel <level>', 'Logging level (default: "info"): "error", "warn", "info", "debug"', 'info')
  .option('-v, --validate', 'Enable validation of OpenAPI spec', false)
  .action((options) => {
    try {
      logger.level = options.loglevel;
      logger.info('Processing OpenAPI spec from stdin');

      let inputData = '';
      process.stdin.on('data', (chunk) => {
        inputData += chunk;
      });

      process.stdin.on('end', () => {
        const openapiSpec = JSON.parse(inputData);

        logger.debug('OpenAPI spec loaded successfully');

        const extractedData = openapiExtractor.extract(openapiSpec, { path: options.pattern, removeDocs: true, removeExamples: true, removeExtensions: true });

        logger.debug('Extracted data:', extractedData);

        outputResult(extractedData, options.output);

        logger.info('Extraction completed successfully');
      });
    } catch (error) {
      logger.error('An error occurred:', error);
      process.exit(1);
    }
  });
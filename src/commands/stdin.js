// commands/stdin.js
import { Command } from 'commander';
import openapiExtractor from 'openapi-extract';
import { logger } from '../utils/logger.js';
import { outputResult } from '../utils/output.js';

export const stdinCommand = new Command('stdin')
  .description('Process OpenAPI spec from stdin')
  .requiredOption('-p, --path <path>', 'The Path to the desired endpoint within the OpenAPI spec (including the leading slash) for example "/auth"')
  .option('-o, --output <output>', 'Output type (default: "json"): "json" or "yaml"', 'json')
  .option('-l, --loglevel <level>', 'Logging level (default: "error"): "error", "warn", "info", "debug"', 'error')
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

        // prefix a leading slash to the options.path if it doesn't already have one
        options.path = options.path.startsWith('/') ? options.path : `/${options.path}`;

        const extractedData = openapiExtractor.extract(openapiSpec, { path: options.path, removeDocs: true, removeExamples: true, removeExtensions: true, openai: true });

        logger.debug('Extracted data:', extractedData);

        outputResult(extractedData, options.output);

        logger.info('Extraction completed successfully');
      });
    } catch (error) {
      logger.error('An error occurred:', error);
      process.exit(1);
    }
  });
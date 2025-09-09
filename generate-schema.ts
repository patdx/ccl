// Script to generate JSON schema from TypeScript schema definition
// Run with: bun run generate-schema.ts

import { toJsonSchema } from '@valibot/to-json-schema';
import * as fs from 'fs';
import schema from './ccl.schema.ts';

const jsonSchema = toJsonSchema(schema);

fs.writeFileSync('./ccl.schema.json', JSON.stringify(jsonSchema, null, 2));

console.log('JSON schema generated successfully!');
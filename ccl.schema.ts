import * as v from 'valibot';

const ConfigSchema = v.object({
  env: v.optional(v.record(v.string(), v.string())),
});

export default v.object({
  $schema: v.nullish(v.string()),
  default: v.nullish(ConfigSchema),
  configs: v.nullish(v.record(v.string(), ConfigSchema)),
})
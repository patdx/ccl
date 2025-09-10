import * as v from 'valibot';

const ConfigSchema = v.pipe(
  v.object({
    env: v.optional(v.pipe(
      v.objectWithRest(
        {
          ANTHROPIC_AUTH_TOKEN: v.optional(v.pipe(
            v.string(),
            v.metadata({
              title: 'Anthropic Auth Token',
              description: 'Authentication token for Anthropic API access.',
              examples: ['sk-ant-api03-...'],
            })
          )),
          ANTHROPIC_BASE_URL: v.optional(v.pipe(
            v.string(),
            v.metadata({
              title: 'Anthropic Base URL',
              description: 'Base URL for Anthropic API endpoints.',
              examples: ['https://api.anthropic.com'],
            })
          )),

        },
        v.string()
      ),
      v.metadata({
        title: 'Environment Variables',
        description: 'Configuration for environment variables with well-known keys and support for additional custom variables.',
        examples: [{ 
          ANTHROPIC_AUTH_TOKEN: 'sk-ant-api03-...', 
          ANTHROPIC_BASE_URL: 'https://api.anthropic.com',
          CUSTOM_VAR: 'custom_value'
        }],
      })
    )),
  }),
  v.metadata({
    title: 'Configuration Schema',
    description: 'Schema defining configuration options for the application.',
    examples: [{ env: { ANTHROPIC_AUTH_TOKEN: 'sk-ant-api03-...', NODE_ENV: 'development' } }],
  })
);

export default v.pipe(
  v.object({
    $schema: v.optional(v.pipe(
      v.string(),
      v.metadata({
        title: 'JSON Schema Reference',
        description: 'Optional reference to the JSON schema for validation.',
        examples: ['./ccl.schema.json'],
      })
    )),
    bin: v.optional(v.pipe(
      v.string(),
      v.metadata({
        title: 'Claude Binary Path',
        description: 'Path to the claude executable. If not provided, will use exec.LookPath to find claude in PATH.',
        examples: ['/home/user/.claude/local/claude', '/usr/local/bin/claude'],
      })
    )),
    default: v.optional(v.pipe(
      ConfigSchema,
      v.metadata({
        title: 'Default Configuration',
        description: 'Default configuration settings applied when no specific config is specified.',
        examples: [{ env: { ANTHROPIC_AUTH_TOKEN: 'sk-ant-api03-...' } }],
      })
    )),
    configs: v.optional(v.pipe(
      v.record(v.string(), ConfigSchema),
      v.metadata({
        title: 'Named Configurations',
        description: 'Named configuration profiles that can be selected at runtime.',
        examples: [{ 
          zai: { env: { ANTHROPIC_BASE_URL: 'https://api.z.ai/api/anthropic', ANTHROPIC_AUTH_TOKEN: 'YOUR_API_KEY' } }
        }],
      })
    )),
  }),
  v.metadata({
    title: 'CCL Configuration Schema',
    description: 'Root schema for CCL (Claude Code Library) configuration files.',
    examples: [{
      $schema: './ccl.schema.json',
      bin: '/home/user/.claude/local/claude',
      default: { env: {} },
      configs: {
        zai: { env: { ANTHROPIC_BASE_URL: 'https://api.z.ai/api/anthropic', ANTHROPIC_AUTH_TOKEN: 'YOUR_API_KEY' } }
      }
    }],
  })
)
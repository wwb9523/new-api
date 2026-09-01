/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
/**
 * Zod schemas for common logs
 * This file should only contain Zod schemas and types inferred from them
 */
import { z } from 'zod'

// Usage log schema
export const usageLogSchema = z.object({
  id: z.number(),
  user_id: z.number(),
  created_at: z.number(),
  type: z.number(),
  content: z.string(),
  username: z.string().default(''),
  token_name: z.string().default(''),
  model_name: z.string().default(''),
  quota: z.number().default(0),
  prompt_tokens: z.number().default(0),
  completion_tokens: z.number().default(0),
  use_time: z.number().default(0),
  is_stream: z.boolean().default(false),
  channel: z.number().default(0),
  channel_name: z.string().nullish().default(''),
  token_id: z.number().default(0),
  group: z.string().default(''),
  ip: z.string().default(''),
  other: z.string().default(''),
  request_id: z.string().default(''),
  upstream_request_id: z.string().default(''),
})

export type UsageLog = z.infer<typeof usageLogSchema>

export const conversationLogSchema = z.object({
  id: z.number(),
  request_id: z.string().default(''),
  user_id: z.number(),
  username: z.string().default(''),
  created_at: z.number(),
  status: z.string().default(''),
  relay_format: z.string().default(''),
  relay_mode: z.number().default(0),
  model_name: z.string().default(''),
  token_name: z.string().default(''),
  token_id: z.number().default(0),
  channel: z.number().default(0),
  channel_name: z.string().nullish().default(''),
  group: z.string().default(''),
  prompt_tokens: z.number().default(0),
  completion_tokens: z.number().default(0),
  quota: z.number().default(0),
  use_time: z.number().default(0),
  is_stream: z.boolean().default(false),
  request_preview: z.string().default(''),
  response_preview: z.string().default(''),
  request_body: z.string().default(''),
  response_body: z.string().default(''),
  error_message: z.string().default(''),
  truncated: z.boolean().default(false),
})

export type ConversationLog = z.infer<typeof conversationLogSchema>

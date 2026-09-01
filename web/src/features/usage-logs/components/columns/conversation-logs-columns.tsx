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
import type { ColumnDef } from '@tanstack/react-table'
import { Eye, KeyRound } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { GroupBadge } from '@/components/group-badge'
import { StatusBadge, type StatusBadgeProps } from '@/components/status-badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { getUserAvatarFallback, getUserAvatarStyle } from '@/lib/avatar'
import {
  formatLogQuota,
  formatTimestampToDate,
  formatUseTime,
} from '@/lib/format'
import { cn } from '@/lib/utils'

import type { ConversationLog } from '../../data/schema'
import { ConversationLogDialog } from '../dialogs/conversation-log-dialog'
import { ModelBadge } from '../model-badge'
import { useUsageLogsContext } from '../usage-logs-provider'

function getConversationStatusLabel(
  status: string,
  t: (key: string) => string
): string {
  if (status === 'success') return t('Success')
  if (status === 'error') return t('Error')
  return status || '-'
}

function getConversationStatusVariant(
  status: string
): StatusBadgeProps['variant'] {
  return status === 'error' ? 'red' : 'green'
}

export function useConversationLogsColumns(
  isAdmin: boolean
): ColumnDef<ConversationLog>[] {
  const { t } = useTranslation()
  const columns: ColumnDef<ConversationLog>[] = [
    {
      accessorKey: 'created_at',
      header: t('Time'),
      cell: ({ row }) => (
        <div className='flex min-w-0 flex-col gap-0.5'>
          <span className='truncate font-mono text-xs tabular-nums'>
            {formatTimestampToDate(row.original.created_at)}
          </span>
          <StatusBadge
            label={getConversationStatusLabel(row.original.status, t)}
            variant={getConversationStatusVariant(row.original.status)}
            size='sm'
            copyable={false}
            className='-ml-1.5 !text-xs [&_span]:!text-xs'
          />
        </div>
      ),
      enableHiding: false,
      size: 170,
    },
  ]

  if (isAdmin) {
    columns.push(
      {
        id: 'channel',
        header: t('Channel'),
        accessorFn: (row) => row.channel,
        cell: ({ row }) => {
          const log = row.original
          if (!log.channel) return null
          return (
            <div className='flex max-w-[160px] flex-col gap-0.5'>
              <StatusBadge
                label={`#${log.channel}`}
                autoColor={String(log.channel)}
                copyText={String(log.channel)}
                size='sm'
                showDot={false}
                className='font-mono'
              />
              {log.channel_name && (
                <span className='text-muted-foreground/70 truncate text-xs'>
                  {log.channel_name}
                </span>
              )}
            </div>
          )
        },
      },
      {
        id: 'username',
        header: t('User'),
        accessorFn: (row) => row.username,
        cell: function UserCell({ row }) {
          const { sensitiveVisible, setSelectedUserId, setUserInfoDialogOpen } =
            useUsageLogsContext()
          const log = row.original
          if (!log.username) return null
          return (
            <button
              type='button'
              className='flex items-center gap-1.5 text-left'
              onClick={(e) => {
                e.stopPropagation()
                setSelectedUserId(log.user_id)
                setUserInfoDialogOpen(true)
              }}
            >
              <Avatar className='ring-border/60 size-6 ring-1 max-sm:hidden'>
                <AvatarFallback
                  className={cn(
                    'text-[11px] font-semibold',
                    !sensitiveVisible && 'bg-muted text-muted-foreground'
                  )}
                  style={
                    sensitiveVisible
                      ? getUserAvatarStyle(log.username)
                      : undefined
                  }
                >
                  {sensitiveVisible ? getUserAvatarFallback(log.username) : '•'}
                </AvatarFallback>
              </Avatar>
              <span className='text-muted-foreground max-w-[110px] truncate text-sm'>
                {sensitiveVisible ? log.username : '••••'}
              </span>
            </button>
          )
        },
      }
    )
  }

  columns.push(
    {
      accessorKey: 'token_name',
      header: t('Token'),
      cell: ({ row }) => {
        const log = row.original
        if (!log.token_name) return null
        return (
          <div className='flex max-w-[190px] flex-col gap-0.5'>
            <StatusBadge
              label={log.token_name}
              icon={KeyRound}
              copyText={log.token_name}
              size='sm'
              showDot={false}
              className='border-border/60 bg-muted/30 text-foreground h-6 max-w-full gap-1.5 overflow-hidden rounded-md border px-2 py-0.5'
            />
            {log.group && (
              <GroupBadge
                group={log.group}
                type='text'
                size='sm'
                className='inline align-baseline text-xs leading-none'
              />
            )}
          </div>
        )
      },
      size: 160,
    },
    {
      accessorKey: 'model_name',
      header: t('Model'),
      cell: ({ row }) => <ModelBadge modelName={row.original.model_name} />,
      meta: { mobileTitle: true },
    },
    {
      accessorKey: 'prompt_tokens',
      header: t('Tokens'),
      cell: ({ row }) => (
        <span className='font-mono text-xs font-medium tabular-nums'>
          {row.original.prompt_tokens.toLocaleString()} /{' '}
          {row.original.completion_tokens.toLocaleString()}
        </span>
      ),
    },
    {
      accessorKey: 'quota',
      header: t('Cost'),
      cell: ({ row }) => formatLogQuota(row.original.quota),
    },
    {
      accessorKey: 'use_time',
      header: t('Timing'),
      cell: ({ row }) => (
        <div className='flex flex-col gap-0.5'>
          <span className='font-mono text-xs'>
            {formatUseTime(row.original.use_time)}
          </span>
          <StatusBadge
            label={row.original.is_stream ? t('Stream') : t('Non-stream')}
            variant={row.original.is_stream ? 'blue' : 'grey'}
            size='sm'
            copyable={false}
          />
        </div>
      ),
    },
    {
      accessorKey: 'request_preview',
      header: t('Request'),
      cell: ({ row }) => (
        <span className='line-clamp-2 max-w-[260px] text-xs break-words'>
          {row.original.request_preview || '-'}
        </span>
      ),
      size: 280,
    },
    {
      accessorKey: 'response_preview',
      header: t('Response'),
      cell: ({ row }) => (
        <span className='line-clamp-2 max-w-[260px] text-xs break-words'>
          {row.original.response_preview || row.original.error_message || '-'}
        </span>
      ),
      size: 280,
    },
    {
      id: 'actions',
      header: '',
      cell: function ConversationActionsCell({ row }) {
        const [open, setOpen] = useState(false)
        return (
          <>
            <Button
              type='button'
              variant='ghost'
              size='icon'
              className='size-7'
              onClick={(e) => {
                e.stopPropagation()
                setOpen(true)
              }}
              aria-label={t('View Details')}
            >
              <Eye className='size-4' />
            </Button>
            <ConversationLogDialog
              log={row.original}
              open={open}
              onOpenChange={setOpen}
              isAdmin={isAdmin}
            />
          </>
        )
      },
      enableHiding: false,
      size: 60,
    }
  )

  return columns
}

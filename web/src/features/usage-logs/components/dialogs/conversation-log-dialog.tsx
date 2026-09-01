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
import { Check, Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { StatusBadge, type StatusBadgeProps } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import {
  formatLogQuota,
  formatTimestampToDate,
  formatUseTime,
} from '@/lib/format'
import { cn } from '@/lib/utils'

import type { ConversationLog } from '../../data/schema'

type ConversationLogDialogProps = {
  log: ConversationLog
  open: boolean
  onOpenChange: (open: boolean) => void
  isAdmin: boolean
}

function statusLabel(status: string, t: (key: string) => string): string {
  if (status === 'success') return t('Success')
  if (status === 'error') return t('Error')
  return status || '-'
}

function statusVariant(status: string): StatusBadgeProps['variant'] {
  return status === 'error' ? 'red' : 'green'
}

function DetailRow(props: {
  label: React.ReactNode
  value: React.ReactNode
  mono?: boolean
}) {
  return (
    <div className='grid min-w-0 grid-cols-[6.5rem_minmax(0,1fr)] gap-3 text-sm'>
      <span className='text-muted-foreground min-w-0 text-xs'>
        {props.label}
      </span>
      <span
        className={cn('min-w-0 break-all text-xs', props.mono && 'font-mono')}
      >
        {props.value}
      </span>
    </div>
  )
}

function BodyBlock(props: {
  title: string
  body: string
  copiedText: string | null
  onCopy: (text: string) => void
}) {
  const { t } = useTranslation()
  const hasBody = props.body.trim() !== ''
  const copied = props.copiedText === props.body

  return (
    <div className='min-w-0 space-y-2'>
      <div className='flex items-center justify-between gap-3'>
        <Label className='text-xs font-semibold'>{props.title}</Label>
        <Button
          type='button'
          variant='ghost'
          size='sm'
          className='h-7 gap-1.5 px-2 text-xs'
          disabled={!hasBody}
          onClick={() => props.onCopy(props.body)}
        >
          {copied ? <Check className='size-3' /> : <Copy className='size-3' />}
          {copied ? t('Copied') : t('Copy')}
        </Button>
      </div>
      <pre className='bg-muted/30 border-border/70 max-h-[260px] min-h-24 overflow-auto rounded-md border p-3 text-xs leading-relaxed break-words whitespace-pre-wrap'>
        {hasBody ? props.body : '-'}
      </pre>
    </div>
  )
}

export function ConversationLogDialog(props: ConversationLogDialogProps) {
  const { t } = useTranslation()
  const { copiedText, copyToClipboard } = useCopyToClipboard({ notify: false })
  const log = props.log

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={
        <>
          {t('Conversation Details')}
          <StatusBadge
            label={statusLabel(log.status, t)}
            variant={statusVariant(log.status)}
            size='sm'
            copyable={false}
          />
        </>
      }
      description={t('View the complete request and response content')}
      contentClassName='sm:max-w-4xl lg:max-w-5xl'
      titleClassName='flex items-center gap-2 text-base'
      descriptionClassName='sr-only'
      contentHeight='min(76dvh, 760px)'
      bodyClassName='pr-2 sm:pr-4'
    >
      <div className='min-w-0 space-y-4'>
        <div className='grid min-w-0 gap-x-6 gap-y-1.5 sm:grid-cols-2'>
          <DetailRow
            label={t('Time')}
            value={formatTimestampToDate(log.created_at)}
          />
          <DetailRow
            label={t('Request ID')}
            value={log.request_id || '-'}
            mono
          />
          {props.isAdmin && (
            <DetailRow
              label={t('User')}
              value={
                log.username ? `${log.username} (${log.user_id})` : log.user_id
              }
            />
          )}
          <DetailRow label={t('Model')} value={log.model_name || '-'} />
          <DetailRow label={t('Token')} value={log.token_name || '-'} />
          <DetailRow label={t('Group')} value={log.group || '-'} />
          {props.isAdmin && (
            <DetailRow
              label={t('Channel')}
              value={
                log.channel_name
                  ? `${log.channel_name} #${log.channel}`
                  : `#${log.channel}`
              }
            />
          )}
          <DetailRow
            label={t('Tokens')}
            value={`${log.prompt_tokens.toLocaleString()} / ${log.completion_tokens.toLocaleString()}`}
          />
          <DetailRow label={t('Cost')} value={formatLogQuota(log.quota)} />
          <DetailRow label={t('Timing')} value={formatUseTime(log.use_time)} />
          <DetailRow label={t('Format')} value={log.relay_format || '-'} />
          <DetailRow
            label={t('Stream')}
            value={log.is_stream ? t('Yes') : t('No')}
          />
          {log.truncated && (
            <DetailRow
              label={t('Storage')}
              value={t('Body content truncated')}
            />
          )}
          {log.error_message && (
            <DetailRow label={t('Error')} value={log.error_message} />
          )}
        </div>

        <BodyBlock
          title={t('Request Body')}
          body={log.request_body}
          copiedText={copiedText}
          onCopy={copyToClipboard}
        />
        <BodyBlock
          title={t('Response Body')}
          body={log.response_body}
          copiedText={copiedText}
          onCopy={copyToClipboard}
        />
      </div>
    </Dialog>
  )
}

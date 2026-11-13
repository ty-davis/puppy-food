import { client } from '../api/generated/client.gen'
import { listFeedings, publishSchedule, setFeedingCompletion } from '../api/generated/sdk.gen'
import type { FeedingFeeding, FeedingFeedingList, FeedingPublishInput } from '../api/generated/types.gen'

client.setConfig({ baseUrl: window.location.origin })

function errorMessage(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'message' in error) {
    return String(error.message)
  }
  return 'Request failed'
}

export async function loadFeedings(from: string, to: string): Promise<FeedingFeeding[]> {
  const { data, error } = await listFeedings({
    query: { from, to },
  })
  if (error) throw new Error(errorMessage(error))
  return (data as FeedingFeedingList | undefined)?.feedings ?? []
}

export async function updateCompletion(id: string, completed: boolean): Promise<void> {
  const { error } = await setFeedingCompletion({
    path: { id },
    body: { completed },
  })
  if (error) throw new Error(errorMessage(error))
}

export async function publish(input: FeedingPublishInput): Promise<void> {
  const { error } = await publishSchedule({
    body: input,
  })
  if (error) throw new Error(errorMessage(error))
}

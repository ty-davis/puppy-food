import assert from 'node:assert/strict'
import test from 'node:test'
import { formatQuantity } from '../src/lib/dates.ts'

test('cup quantities round to eighths and reduce the fraction', () => {
  const examples = [
    [0.125, '1/8'],
    [0.25, '1/4'],
    [0.375, '3/8'],
    [0.5, '1/2'],
    [0.75, '3/4'],
    [0.875, '7/8'],
    [1, '1'],
    [1.125, '1 1/8'],
  ]

  for (const [quantity, expected] of examples) {
    assert.equal(formatQuantity(quantity, 'c'), expected)
  }
})

test('cup quantities round to the nearest eighth and accept a capital C', () => {
  assert.equal(formatQuantity(0.19, 'C'), '1/4')
  assert.equal(formatQuantity(0.06, 'c'), '0')
})

test('non-cup quantities retain decimal formatting', () => {
  assert.equal(formatQuantity(0.3333, 'g'), '0.333')
})

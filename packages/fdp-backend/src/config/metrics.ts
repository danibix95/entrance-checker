
import PrometheusClient, { Counter } from 'prom-client'

export const getMetrics = (prometheusClient: typeof PrometheusClient): {
  // add here your exported metrics type
  counterExample: Counter<any>
} => {
  // define here your custom metrics
  const counterExample = new prometheusClient.Counter({
    name: 'example_total',
    help: 'count how many times this metric is incremented',
    labelNames: [],
  })

  return {
    counterExample,
  }
}

export type Metrics = ReturnType<typeof getMetrics>

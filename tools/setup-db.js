const { Client } = require('pg')

async function setup() {
  // Try different connection options
  const configs = [
    { host: 'localhost', user: 'postgres', database: 'postgres', port: 5432 },
    { host: '127.0.0.1', user: 'postgres', database: 'postgres', port: 5432 },
  ]

  for (const config of configs) {
    const client = new Client(config)
    try {
      await client.connect()
      console.log('Connected with:', config)

      // Try to create database
      try {
        await client.query('CREATE DATABASE secretflow')
        console.log('Database secretflow created!')
      } catch (err) {
        if (err.code === '42P04') {
          console.log('Database secretflow already exists')
        } else {
          console.error('Error creating database:', err.message)
        }
      }

      // Apply migrations
      const fs = require('fs')
      const migration = fs.readFileSync('../backend/migrations/001_initial_schema.sql', 'utf8')

      const client2 = new Client({ ...config, database: 'secretflow' })
      await client2.connect()
      await client2.query(migration)
      console.log('Migrations applied!')
      await client2.end()

      await client.end()
      console.log('Setup complete!')
      process.exit(0)
    } catch (err) {
      console.error('Failed with config:', config, err.message)
      await client.end().catch(() => {})
    }
  }

  console.log('All connection attempts failed')
  process.exit(1)
}

setup()

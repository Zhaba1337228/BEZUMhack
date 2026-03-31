const { Client } = require('pg')

async function setup() {
  // Connect as current Windows user via SSPI/SSPI trust
  const configs = [
    { host: 'localhost', database: 'postgres', user: 'vasak' },
    { host: '127.0.0.1', database: 'postgres', user: 'vasak' },
  ]

  for (const config of configs) {
    const client = new Client(config)
    try {
      await client.connect()
      console.log('Connected as:', config.user)

      // Create role secretflow if not exists
      try {
        await client.query("CREATE ROLE secretflow WITH LOGIN PASSWORD 'secretflow_password' CREATEDB;")
        console.log('Role secretflow created!')
      } catch (err) {
        if (err.code === '42710') {
          console.log('Role secretflow already exists')
        } else {
          console.error('Error creating role:', err.message)
        }
      }

      // Grant privileges
      await client.query('GRANT ALL PRIVILEGES ON DATABASE secretflow TO secretflow;')
      console.log('Privileges granted!')

      // Connect to secretflow and grant schema privileges
      const client2 = new Client({ ...config, database: 'secretflow' })
      await client2.connect()
      await client2.query('GRANT ALL ON SCHEMA public TO secretflow;')
      console.log('Schema privileges granted!')
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

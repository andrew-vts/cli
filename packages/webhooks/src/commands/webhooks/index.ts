import {Command, flags} from '@heroku-cli/command'

export default class WebhooksIndex extends Command {
  static description = 'describe the command here'

  static flags = {
    app: flags.app(),
    remote: flags.remote(),
    pipeline: flags.string({ char: 'p', description: 'pipeline on which to list', hidden: true })
  }

  static args = [{name: 'file'}]

  async run() {
    const {args, flags} = this.parse(WebhooksIndex)

    const name = flags.name || 'world'
    this.log(`hello ${name} from /Users/ccarbert/projects/heroku/cli/packages/webhooks/src/commands/webhooks/index.ts`)
    if (args.file && flags.force) {
      this.log(`you input --force and --file: ${args.file}`)
    }
  }
}

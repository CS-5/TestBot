import * as fs from 'fs';
import * as Discord from 'discord.js';
import * as dotenv from 'dotenv';
import { Command } from './models';
import command from './commands/ping';

const client = new Discord.Client();
const commands: Discord.Collection<string, Command> = new Discord.Collection();
dotenv.config();

const PREFIX = process.env.MESSAGE_PREFIX;

if (!PREFIX) {
  console.warn(
    `Message prefix not found. Please add an environment variable "MESSAGE_PREFIX", with the desired prefix. (default is '!')`
  );
  process.exit();
}

const commandFiles = fs
  .readdirSync('./src/commands')
  .filter(file => file.endsWith('.ts'));

commandFiles.forEach(async filename => {
  const { default: command } = await import('./commands/' + filename);
  commands!.set(command.name, command);
});

client.on('ready', () => {
  console.log('Bot connected');
});

client.on('message', message => {
  if (!commands) {
    console.warn(`No commands founds`);
  }
  if (!message.content.startsWith(PREFIX as string) || message.author.bot) {
    return;
  }

  const args = message.content.slice(PREFIX!.length).split(/ +/);
  const command = args.shift()!.toLowerCase();

  if (!commands.has(command)) {
    return;
  }

  try {
    const cmd = commands.get(command);
    cmd && cmd.execute(message, args);
  } catch (error) {
    console.error(error);
    message.reply('There was an error while trying to execute that command');
  }
});

client.login(process.env.API_KEY); // haha get pranked API key scraper!

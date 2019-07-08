import * as fs from 'fs';
import * as Discord from 'discord.js';
import * as dotenv from 'dotenv';
import { Command } from './models';

dotenv.config();

const client = new Discord.Client();
const commands: Discord.Collection<string, Command> = new Discord.Collection();
const PREFIX = process.env.MESSAGE_PREFIX;
const APIKEY = process.env.API_KEY;

// Check for prefix env var
if (!PREFIX) {
  console.warn(
    `Message prefix not found. Please add an environment variable "MESSAGE_PREFIX", with the desired prefix. (default is '!')`
  );
  process.exit();
}

// Build list of command files
const commandFiles = fs
  .readdirSync('./src/commands')
  .filter(file => file.endsWith('.ts'));

// Add commands to the collection
commandFiles.forEach(async filename => {
  const { default: command } = await import('./commands/' + filename);
  commands!.set(command.name, command);
});

// Inform the operator that the bot is connected
client.on('ready', () => {
  console.log('Bot connected');
});

// Handle a message being recieved
client.on('message', message => {
  // If there are no commands, inform the operator
  if (!commands) {
    console.warn(`No commands founds`);
  }

  // Ignore all messages without the prefix or those sent by a bot
  if (!message.content.startsWith(PREFIX as string) || message.author.bot) {
    return;
  }

  // Build variables based on the command executed
  const args = message.content.slice(PREFIX!.length).split(/ +/);
  const command = args.shift()!.toLowerCase();

  // If the command does not exist, return
  if (!commands.has(command)) {
    return;
  }

  // Attempt to execute the command
  try {
    const cmd = commands.get(command);
    cmd && cmd.execute(message, args);
  } catch (error) {
    console.error(error);
    message.reply('There was an error while trying to execute that command');
  }
});

client.login(APIKEY); // haha get pranked API key scraper!

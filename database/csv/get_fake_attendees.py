#! /usr/bin/env python

import numpy as np
import pandas as pd
import datetime

np.random.seed(42)

# MODEL STRUCTURE
class Attendee:
  def __init__(self, ticket_num, last_name = None, first_name = None, ticket_type = 10, vendor = None, resp_vendor = None):
    self.ticket_num = ticket_num
    self.last_name = last_name.capitalize() if isinstance(last_name, str) else last_name
    self.first_name = first_name.capitalize() if isinstance(first_name, str) else first_name
    self.ticket_type = ticket_type if ticket_num < 1000 else 0
    self.sold = 1 if last_name is not None and np.random.random() > 0.15 else 0
    self.vendor = vendor if last_name is not None else None
    self.resp_vendor = resp_vendor
    self.entered = datetime.datetime.now().isoformat() if last_name is not None and np.random.random() > 0.7 else None

  def to_dict(self):
    return {
      'ticket_num': self.ticket_num,
      'last_name': self.last_name,
      'first_name': self.first_name,
      'ticket_type': self.ticket_type,
      'sold': self.sold,
      'vendor': self.vendor,
      'resp_vendor': self.resp_vendor,
      'entered': self.entered,
    }

  def __lt__(self, other):
         return self.ticket_num < other.ticket_num


# how many tickets should be sold for the test
num_tickets = 800
total_tickets = 1100

# generate vendors
vendors = pd.read_csv('https://randomuser.me/api/?results={}&inc=name&nat=nl,es,de&format=csv&seed=42'.format(20),
                      names=['title', 'first_name', 'last_name'], header=0)
vendors = vendors.drop('title', axis=1)
vendors_gen = [' '.join(list(v)[1:]).title() for v in vendors.to_records()]

# generate attendees
attendees = pd.read_csv('https://randomuser.me/api/?results={}&inc=name&nat=nl,es,de&format=csv&seed=40320'.format(num_tickets),
                        names=['title', 'first_name', 'last_name'], header=0)
attendees = attendees.drop('title', axis=1)

ticket_numbers = np.random.choice(total_tickets, total_tickets, replace=False)
attendees_gen = [Attendee(i+1, r[2], r[1], 10, vendors_gen[np.random.randint(len(vendors_gen))], 'Generator')
                 for r,i in zip(attendees.to_records(), ticket_numbers[:attendees.shape[0]])]

# add unsold tickets and sort the results according to ticket number
attendees_gen += [Attendee(t+1, resp_vendor='Generator') for t in ticket_numbers[attendees.shape[0]:]]
attendees_gen.sort()

# write data into proprer csv file
final_df = pd.DataFrame.from_records([a.to_dict() for a in attendees_gen])
final_df.to_csv('attendees.csv', na_rep='null', header=False,
                index=False, columns=['ticket_num','last_name','first_name',
                                      'ticket_type','sold','vendor','resp_vendor','entered'])
